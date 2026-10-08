package fwtypes

import (
	"context"
	"time"

	"github.com/anthropics/anthropic-sdk-go/packages/respjson"
	"github.com/hashicorp/terraform-plugin-framework/types"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type option struct {
	unknownAsNull bool
}

type Option func(*option)

func WithUnknownAsNull() Option {
	return func(opt *option) {
		opt.unknownAsNull = true
	}
}

func StringValue[T ~string](value T, meta respjson.Field, opts ...Option) types.String {
	opt := &option{}
	for _, optFunc := range opts {
		optFunc(opt)
	}

	if !meta.Valid() {
		if meta.Raw() == respjson.Null || opt.unknownAsNull {
			return types.StringNull()
		}
		return types.StringUnknown()
	}
	return types.StringValue(string(value))
}

func BoolValue[T ~bool](value T, meta respjson.Field, opts ...Option) types.Bool {
	opt := &option{}
	for _, optFunc := range opts {
		optFunc(opt)
	}

	if !meta.Valid() {
		if meta.Raw() == respjson.Null || opt.unknownAsNull {
			return types.BoolNull()
		}
		return types.BoolUnknown()
	}
	return types.BoolValue(bool(value))
}

func Int64Value[T ~int64](value T, meta respjson.Field, opts ...Option) types.Int64 {
	opt := &option{}
	for _, optFunc := range opts {
		optFunc(opt)
	}

	if !meta.Valid() {
		if meta.Raw() == respjson.Null || opt.unknownAsNull {
			return types.Int64Null()
		}
		return types.Int64Unknown()
	}
	return types.Int64Value(int64(value))
}

func TimeValue(value time.Time, meta respjson.Field, opts ...Option) types.String {
	return StringValue(value.Format(time.RFC3339Nano), meta, opts...)
}

func ListStringValue(ctx context.Context, value []string, meta respjson.Field, opts ...Option) supertypes.ListValueOf[string] {
	opt := &option{}
	for _, optFunc := range opts {
		optFunc(opt)
	}

	if !meta.Valid() {
		if meta.Raw() == respjson.Null || opt.unknownAsNull {
			return supertypes.NewListValueOfNull[string](ctx)
		}
		return supertypes.NewListValueOfUnknown[string](ctx)
	}
	return supertypes.NewListValueOfSlice(ctx, value)
}
