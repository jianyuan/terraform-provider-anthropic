package fwtypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

func NullableBoolValue(value nullable.Nullable[bool]) types.Bool {
	if v, err := value.Get(); err == nil {
		return types.BoolValue(v)
	} else {
		return types.BoolNull()
	}
}
