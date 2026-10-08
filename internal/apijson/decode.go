package apijson

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"reflect"
	"strings"
	"time"

	"github.com/anthropics/anthropic-sdk-go/packages/respjson"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type Option func(*decoder)

// WithComputedOnly fills only fields tagged `computed` or `computed_optional`,
// leaving user-configured attributes untouched so the resulting state stays
// consistent with the plan.
func WithComputedOnly() Option {
	return func(d *decoder) {
		d.computedOnly = true
	}
}

// WithUnknownAsNull maps "field missing from the response" to null instead of unknown.
func WithUnknownAsNull() Option {
	return func(d *decoder) { d.unknownAsNull = true }
}

// Decode copies every tagged field of src (an API struct or pointer to one)
// into dst (a pointer to a model struct).
func Decode(ctx context.Context, src, dst any, opts ...Option) error {
	return decode(ctx, src, dst, opts)
}

// DecodeComputed only fills model fields tagged `computed` or
// `computed_optional`, leaving user-configured attributes untouched so the
// resulting state stays consistent with the plan.
//
//   - computed:          always overwritten from src
//   - computed_optional: only filled if the current value is null or unknown
//     (i.e. the user did not configure it)
func DecodeComputed(ctx context.Context, src, dst any, opts ...Option) error {
	return decode(ctx, src, dst, append(opts, WithComputedOnly()))
}

func decode(ctx context.Context, src, dst any, opts []Option) error {
	d := &decoder{ctx: ctx}
	for _, o := range opts {
		o(d)
	}

	dv := reflect.ValueOf(dst)
	if dv.Kind() != reflect.Pointer || dv.IsNil() || dv.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("apijson: dst must be a non-nil pointer to a struct, got %T", dst)
	}
	sv, ok := indirect(reflect.ValueOf(src))
	if !ok {
		return fmt.Errorf("apijson: src must not be nil, got %T", src)
	}

	switch sv.Kind() {
	case reflect.Struct:
		return d.decodeStruct(sv, dv.Elem(), "")
	case reflect.Slice, reflect.Array, reflect.Map:
		return d.decodeWhole(sv, dv.Elem())
	}
	return fmt.Errorf("apijson: src must be a struct, slice or map (or pointer to one), got %T", src)
}

// Field presence is whether the JSON field was present, null, or unknown.

type fieldPresence uint8

const (
	fieldPresent fieldPresence = iota
	fieldNull
	fieldUnknown
)

func stateOfField(meta respjson.Field, hasMeta bool) fieldPresence {
	if !hasMeta || meta.Valid() {
		return fieldPresent
	}
	if meta.Raw() == respjson.Null {
		return fieldNull
	}
	return fieldUnknown
}

// Struct-level decoding

type fieldTag struct {
	name             string
	computed         bool
	computedOptional bool
}

func (t fieldTag) isComputed() bool { return t.computed || t.computedOptional }

func parseField(f reflect.StructField) (fieldTag, bool) {
	tf := f.Tag.Get("tfsdk")
	if tf == "" || tf == "-" {
		return fieldTag{}, false
	}
	t := fieldTag{name: tf}
	if raw, ok := f.Tag.Lookup("apijson"); ok {
		if raw == "-" {
			return fieldTag{}, false
		}
		name, flags, _ := strings.Cut(raw, ",")
		if name != "" {
			t.name = name
		}
		for _, fl := range strings.Split(flags, ",") {
			switch strings.TrimSpace(fl) {
			case "computed":
				t.computed = true
			case "computed_optional":
				t.computedOptional = true
			}
		}
	}
	return t, true
}

// indexSource maps json tag name -> struct field index for the API struct.
func indexSource(t reflect.Type) map[string]int {
	out := make(map[string]int, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		name := f.Name
		if tag, ok := f.Tag.Lookup("json"); ok {
			if tag == "-" {
				continue
			}
			if n, _, _ := strings.Cut(tag, ","); n != "" {
				name = n
			}
		}
		out[name] = i
	}
	return out
}

// lookupMeta finds src.JSON.<goFieldName> as a respjson.Field.
func lookupMeta(jsonStruct reflect.Value, goName string) (respjson.Field, bool) {
	if !jsonStruct.IsValid() || jsonStruct.Kind() != reflect.Struct {
		return respjson.Field{}, false
	}
	f := jsonStruct.FieldByName(goName)
	if !f.IsValid() || !f.CanInterface() {
		return respjson.Field{}, false
	}
	m, ok := f.Interface().(respjson.Field)
	return m, ok
}

func stateOf(meta respjson.Field, hasMeta bool) fieldPresence {
	if !hasMeta || meta.Valid() {
		return fieldPresent
	}
	if meta.Raw() == respjson.Null {
		return fieldNull
	}
	return fieldUnknown
}

type decoder struct {
	ctx           context.Context
	computedOnly  bool
	unknownAsNull bool
}

// nestedObjectType is the subset of supertypes.NestedObjectType we rely on. It
// is declared locally so this package doesn't import supertypes.
type nestedObjectType interface {
	attr.Type
	NewObjectPtr(context.Context) (any, diag.Diagnostics)
	NullValue(context.Context) (attr.Value, diag.Diagnostics)
	ValueFromObjectPtr(context.Context, any) (attr.Value, diag.Diagnostics)
	ValueFromObjectSlice(context.Context, any) (attr.Value, diag.Diagnostics)
}

var (
	attrValueType = reflect.TypeFor[attr.Value]()
	timeType      = reflect.TypeFor[time.Time]()

	listT = reflect.TypeFor[types.List]()
	setT  = reflect.TypeFor[types.Set]()
	mapT  = reflect.TypeFor[types.Map]()
)

func (d *decoder) decodeStruct(src, dst reflect.Value, path string) error {
	srcType := src.Type()
	srcIndex := indexSource(srcType)
	jsonMeta := src.FieldByName("JSON")
	for i := 0; i < dst.NumField(); i++ {
		sf := dst.Type().Field(i)
		if !sf.IsExported() {
			continue
		}
		tag, ok := parseField(sf)
		if !ok {
			continue
		}
		if d.computedOnly && !tag.isComputed() {
			continue
		}

		fieldPath := join(path, tag.name)
		dstField := dst.Field(i)

		if d.computedOnly && tag.computedOptional && dstField.Type().Implements(attrValueType) {
			if v, ok := dstField.Interface().(attr.Value); ok && !v.IsNull() && !v.IsUnknown() {
				continue // user configured it; keep their value
			}
		}

		idx, ok := srcIndex[tag.name]
		if !ok {
			return fmt.Errorf("apijson: %s: no field with json name %q in %s (tag the model field `apijson:\"-\"` to skip it)",
				fieldPath, tag.name, srcType)
		}

		meta, hasMeta := lookupMeta(jsonMeta, srcType.Field(idx).Name)
		if err := d.assign(dstField, src.Field(idx), stateOfField(meta, hasMeta), fieldPath); err != nil {
			return err
		}
	}
	return nil
}

func (d *decoder) decodeWhole(src, dst reflect.Value) error {
	var targets []int
	var tags []fieldTag
	for i := 0; i < dst.NumField(); i++ {
		sf := dst.Type().Field(i)
		if !sf.IsExported() {
			continue
		}
		t, ok := parseField(sf)
		if !ok || !t.isComputed() || !d.isCollectionFor(sf.Type, src.Kind()) {
			continue
		}
		targets = append(targets, i)
		tags = append(tags, t)
	}

	switch len(targets) {
	case 0:
		return fmt.Errorf("apijson: src is %s, but %s has no computed collection field to receive it (tag it `apijson:\",computed\"`)",
			src.Type(), dst.Type())
	case 1:
	default:
		names := make([]string, len(tags))
		for i, t := range tags {
			names[i] = t.name
		}
		return fmt.Errorf("apijson: src is %s, but %s has several computed collection fields (%s); keep exactly one or tag the others `apijson:\"-\"`",
			src.Type(), dst.Type(), strings.Join(names, ", "))
	}

	field := dst.Field(targets[0])
	if d.computedOnly && tags[0].computedOptional {
		if v, ok := field.Interface().(attr.Value); ok && !v.IsNull() && !v.IsUnknown() {
			return nil // user configured it; keep their value
		}
	}
	return d.assign(field, src, fieldPresent, tags[0].name)
}

// isCollectionFor reports whether a model field of Go type ft can receive a
// source value of the given kind.
func (d *decoder) isCollectionFor(ft reflect.Type, k reflect.Kind) bool {
	if !ft.Implements(attrValueType) {
		return false
	}
	zero, ok := reflect.Zero(ft).Interface().(attr.Value)
	if !ok {
		return false
	}
	switch zero.Type(d.ctx).TerraformType(d.ctx).(type) {
	case tftypes.List, tftypes.Set:
		return k == reflect.Slice || k == reflect.Array
	case tftypes.Map:
		return k == reflect.Map
	}
	return false
}

func (d *decoder) assign(dst, src reflect.Value, st fieldPresence, path string) error {
	dt := dst.Type()
	switch {
	case dt.Implements(attrValueType):
		return d.assignValue(dst, src, st, path)

	case dt.Kind() == reflect.Pointer && dt.Elem().Kind() == reflect.Struct:
		s, ok := indirect(src)
		if st != fieldPresent || !ok {
			dst.Set(reflect.Zero(dt))
			return nil
		}
		if s.Kind() != reflect.Struct {
			return fmt.Errorf("apijson: %s: cannot decode %s into %s", path, s.Type(), dt)
		}
		n := reflect.New(dt.Elem())
		if err := d.decodeStruct(s, n.Elem(), path); err != nil {
			return err
		}
		dst.Set(n)
		return nil

	case dt.Kind() == reflect.Struct:
		s, ok := indirect(src)
		if st != fieldPresent || !ok {
			dst.Set(reflect.Zero(dt))
			return nil
		}
		if s.Kind() != reflect.Struct {
			return fmt.Errorf("apijson: %s: cannot decode %s into %s", path, s.Type(), dt)
		}
		return d.decodeStruct(s, dst, path)
	}
	return fmt.Errorf("apijson: %s: unsupported model field type %s", path, dt)
}

func (d *decoder) assignValue(dst, src reflect.Value, st fieldPresence, path string) error {
	t, err := d.attrTypeFor(dst.Type(), derefType(src.Type()))
	if err != nil {
		return fmt.Errorf("apijson: %s: %w", path, err)
	}

	var v attr.Value
	switch nt, isNested := t.(nestedObjectType); {
	case st != fieldPresent:
		v, err = d.special(t, st)
	case isNested:
		v, err = d.nested(nt, src, path)
	default:
		v, err = d.value(t, src, path)
	}
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if !rv.IsValid() || !rv.Type().AssignableTo(dst.Type()) {
		return fmt.Errorf("apijson: %s: cannot assign %T to %s", path, v, dst.Type())
	}
	dst.Set(rv)
	return nil
}

// attrTypeFor returns the attr.Type to build for a model field of Go type dst.
// The bare framework collection types carry no element type, so for those it
// is inferred from the API struct's Go type; every other attr.Value is asked
// for its own Type().
func (d *decoder) attrTypeFor(dst, src reflect.Type) (attr.Type, error) {
	switch dst {
	case listT, setT:
		if src.Kind() != reflect.Slice && src.Kind() != reflect.Array {
			return nil, fmt.Errorf("cannot decode %s into %s", src, dst)
		}
		et, err := attrTypeForSrc(src.Elem())
		if err != nil {
			return nil, err
		}
		if dst == listT {
			return basetypes.ListType{ElemType: et}, nil
		}
		return basetypes.SetType{ElemType: et}, nil
	case mapT:
		if src.Kind() != reflect.Map || src.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("cannot decode %s into %s", src, dst)
		}
		et, err := attrTypeForSrc(src.Elem())
		if err != nil {
			return nil, err
		}
		return basetypes.MapType{ElemType: et}, nil
	}

	zero, ok := reflect.Zero(dst).Interface().(attr.Value)
	if !ok {
		return nil, fmt.Errorf("%s is not an attr.Value", dst)
	}
	return zero.Type(d.ctx), nil
}

func attrTypeForSrc(t reflect.Type) (attr.Type, error) {
	t = derefType(t)
	if t == timeType {
		return types.StringType, nil
	}
	switch t.Kind() {
	case reflect.String:
		return types.StringType, nil
	case reflect.Bool:
		return types.BoolType, nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return types.Int64Type, nil
	case reflect.Float32, reflect.Float64:
		return types.Float64Type, nil
	case reflect.Slice, reflect.Array:
		et, err := attrTypeForSrc(t.Elem())
		if err != nil {
			return nil, err
		}
		return basetypes.ListType{ElemType: et}, nil
	case reflect.Map:
		if t.Key().Kind() != reflect.String {
			return nil, fmt.Errorf("unsupported map key type %s", t.Key())
		}
		et, err := attrTypeForSrc(t.Elem())
		if err != nil {
			return nil, err
		}
		return basetypes.MapType{ElemType: et}, nil
	case reflect.Struct:
		return nil, fmt.Errorf("cannot infer an attribute type for struct %s; use a nested object type such as supertypes.SetNestedObjectValueOf[T]", t)
	}
	return nil, fmt.Errorf("unsupported API field type %s", t)
}

// special builds a null or unknown value of type t.
func (d *decoder) special(t attr.Type, st fieldPresence) (attr.Value, error) {
	var raw any // nil => null
	if st == fieldUnknown && !d.unknownAsNull {
		raw = tftypes.UnknownValue
	}
	return t.ValueFromTerraform(d.ctx, tftypes.NewValue(t.TerraformType(d.ctx), raw))
}

// value converts a present (valid) src into an attr.Value of type t.
func (d *decoder) value(t attr.Type, src reflect.Value, path string) (attr.Value, error) {
	tv, err := d.tfValue(t, src, path)
	if err != nil {
		return nil, err
	}
	v, err := t.ValueFromTerraform(d.ctx, tv)
	if err != nil {
		return nil, fmt.Errorf("apijson: %s: %w", path, err)
	}
	return v, nil
}

func (d *decoder) tfValue(t attr.Type, src reflect.Value, path string) (tftypes.Value, error) {
	tfT := t.TerraformType(d.ctx)

	src, ok := indirect(src)
	if !ok {
		return tftypes.NewValue(tfT, nil), nil
	}
	mismatch := func() (tftypes.Value, error) {
		return tftypes.Value{}, fmt.Errorf("apijson: %s: cannot convert %s to %s", path, src.Type(), t)
	}

	switch {
	case tfT.Is(tftypes.String):
		s, ok := stringOf(src)
		if !ok {
			return mismatch()
		}
		return tftypes.NewValue(tfT, s), nil

	case tfT.Is(tftypes.Bool):
		if src.Kind() != reflect.Bool {
			return mismatch()
		}
		return tftypes.NewValue(tfT, src.Bool()), nil

	case tfT.Is(tftypes.Number):
		f, err := numberOf(src)
		if err != nil {
			return tftypes.Value{}, fmt.Errorf("apijson: %s: %w", path, err)
		}
		if f == nil {
			return mismatch()
		}
		return tftypes.NewValue(tfT, f), nil
	}

	switch tfT.(type) {
	case tftypes.List, tftypes.Set:
		et, ok := t.(attr.TypeWithElementType)
		if !ok || (src.Kind() != reflect.Slice && src.Kind() != reflect.Array) {
			return mismatch()
		}
		elems := make([]tftypes.Value, 0, src.Len())
		for i := 0; i < src.Len(); i++ {
			ev, err := d.tfValue(et.ElementType(), src.Index(i), fmt.Sprintf("%s[%d]", path, i))
			if err != nil {
				return tftypes.Value{}, err
			}
			elems = append(elems, ev)
		}
		return tftypes.NewValue(tfT, elems), nil

	case tftypes.Map:
		et, ok := t.(attr.TypeWithElementType)
		if !ok || src.Kind() != reflect.Map || src.Type().Key().Kind() != reflect.String {
			return mismatch()
		}
		elems := make(map[string]tftypes.Value, src.Len())
		for it := src.MapRange(); it.Next(); {
			k := it.Key().String()
			ev, err := d.tfValue(et.ElementType(), it.Value(), path+"."+k)
			if err != nil {
				return tftypes.Value{}, err
			}
			elems[k] = ev
		}
		return tftypes.NewValue(tfT, elems), nil

	case tftypes.Object:
		at, ok := t.(attr.TypeWithAttributeTypes)
		if !ok || src.Kind() != reflect.Struct {
			return mismatch()
		}
		srcIndex := indexSource(src.Type())
		jsonMeta := src.FieldByName("JSON")
		attrs := at.AttributeTypes()
		vals := make(map[string]tftypes.Value, len(attrs))
		for name, atype := range attrs {
			atfT := atype.TerraformType(d.ctx)
			idx, found := srcIndex[name]
			if !found {
				vals[name] = tftypes.NewValue(atfT, nil)
				continue
			}
			meta, has := lookupMeta(jsonMeta, src.Type().Field(idx).Name)
			switch stateOf(meta, has) {
			case fieldNull:
				vals[name] = tftypes.NewValue(atfT, nil)
			case fieldUnknown:
				if d.unknownAsNull {
					vals[name] = tftypes.NewValue(atfT, nil)
				} else {
					vals[name] = tftypes.NewValue(atfT, tftypes.UnknownValue)
				}
			default:
				v, err := d.tfValue(atype, src.Field(idx), join(path, name))
				if err != nil {
					return tftypes.Value{}, err
				}
				vals[name] = v
			}
		}
		return tftypes.NewValue(tfT, vals), nil
	}
	return mismatch()
}

// nested builds a value for nested-object types (supertypes.*NestedObjectValueOf[T]).
// The element model T is discovered through NewObjectPtr and decoded with its
// own tags; every field of T is decoded, regardless of computedOnly.
func (d *decoder) nested(nt nestedObjectType, src reflect.Value, path string) (attr.Value, error) {
	src, ok := indirect(src)
	if !ok {
		v, diags := nt.NullValue(d.ctx)
		if err := diagErr(diags); err != nil {
			return nil, fmt.Errorf("apijson: %s: %w", path, err)
		}
		return v, nil
	}

	proto, diags := nt.NewObjectPtr(d.ctx)
	if err := diagErr(diags); err != nil {
		return nil, fmt.Errorf("apijson: %s: %w", path, err)
	}
	ptrT := reflect.TypeOf(proto)
	if ptrT == nil || ptrT.Kind() != reflect.Pointer || ptrT.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("apijson: %s: nested object type %s did not return a *struct (got %T)", path, nt, proto)
	}

	child := *d
	child.computedOnly = false

	var (
		v attr.Value
		e diag.Diagnostics
	)
	switch src.Kind() {
	case reflect.Slice, reflect.Array:
		out := reflect.MakeSlice(reflect.SliceOf(ptrT), 0, src.Len())
		for i := 0; i < src.Len(); i++ {
			elem, ok := indirect(src.Index(i))
			if !ok {
				continue
			}
			if elem.Kind() != reflect.Struct {
				return nil, fmt.Errorf("apijson: %s[%d]: expected struct, got %s", path, i, elem.Type())
			}
			p := reflect.New(ptrT.Elem())
			if err := child.decodeStruct(elem, p.Elem(), fmt.Sprintf("%s[%d]", path, i)); err != nil {
				return nil, err
			}
			out = reflect.Append(out, p)
		}
		v, e = nt.ValueFromObjectSlice(d.ctx, out.Interface())

	case reflect.Struct:
		p := reflect.New(ptrT.Elem())
		if err := child.decodeStruct(src, p.Elem(), path); err != nil {
			return nil, err
		}
		v, e = nt.ValueFromObjectPtr(d.ctx, p.Interface())

	default:
		return nil, fmt.Errorf("apijson: %s: cannot decode %s into %s", path, src.Type(), nt)
	}
	if err := diagErr(e); err != nil {
		return nil, fmt.Errorf("apijson: %s: %w", path, err)
	}
	return v, nil
}

func stringOf(v reflect.Value) (string, bool) {
	if v.Type() == timeType {
		return v.Interface().(time.Time).Format(time.RFC3339Nano), true
	}
	if v.Kind() == reflect.String {
		return v.String(), true
	}
	return "", false
}

// numberOf returns (nil, nil) when v is not numeric.
func numberOf(v reflect.Value) (*big.Float, error) {
	switch v.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return new(big.Float).SetInt64(v.Int()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return new(big.Float).SetUint64(v.Uint()), nil
	case reflect.Float32, reflect.Float64:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil, fmt.Errorf("cannot represent %v as a number", f)
		}
		return new(big.Float).SetFloat64(f), nil
	}
	return nil, nil
}

func indirect(v reflect.Value) (reflect.Value, bool) {
	for v.IsValid() && (v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface) {
		if v.IsNil() {
			return v, false
		}
		v = v.Elem()
	}
	return v, v.IsValid()
}

func derefType(t reflect.Type) reflect.Type {
	for t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t
}

func join(path, name string) string {
	if path == "" {
		return name
	}
	return path + "." + name
}

func diagErr(diags diag.Diagnostics) error {
	if !diags.HasError() {
		return nil
	}
	var errs []error
	for _, e := range diags.Errors() {
		errs = append(errs, fmt.Errorf("%s: %s", e.Summary(), e.Detail()))
	}
	return errors.Join(errs...)
}
