package provider

import (
	"fmt"
	"strings"
)

func BuildTwoPartId(a, b string) string {
	return fmt.Sprintf("%s/%s", a, b)
}

func SplitTwoPartId(id, a, b string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected %s/%s", id, a, b)
	}
	return parts[0], parts[1], nil
}
