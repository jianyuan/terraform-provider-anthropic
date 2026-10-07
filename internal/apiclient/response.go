package apiclient

import (
	"errors"
	"net/http"

	"github.com/anthropics/anthropic-sdk-go"
)

func IsNotFoundError(err error) bool {
	if apiErr, ok := errors.AsType[*anthropic.Error](err); ok {
		return apiErr.StatusCode == http.StatusNotFound
	}
	return false
}
