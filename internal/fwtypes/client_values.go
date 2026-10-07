package fwtypes

import (
	"time"

	"github.com/anthropics/anthropic-sdk-go/packages/respjson"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringValue[T ~string](value T, meta respjson.Field) types.String {
	if !meta.Valid() {
		if meta.Raw() == respjson.Null {
			return types.StringNull()
		}
		return types.StringUnknown()
	}
	return types.StringValue(string(value))
}

func BoolValue[T ~bool](value T, meta respjson.Field) types.Bool {
	if !meta.Valid() {
		if meta.Raw() == respjson.Null {
			return types.BoolNull()
		}
		return types.BoolUnknown()
	}
	return types.BoolValue(bool(value))
}

func TimeValue(value time.Time, meta respjson.Field) types.String {
	return StringValue(value.Format(time.RFC3339Nano), meta)
}
