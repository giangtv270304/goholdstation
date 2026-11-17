package goholdstation

import (
	"net/url"
	"strconv"
	"strings"
)

type request struct {
	Method string
	Path   string
	Query  url.Values
	Body   any

	Signed               bool
	AllowBusinessFailure bool
	DataKeyInPayload     bool
}

func setPositive(query url.Values, key string, value int64) {
	if value > 0 {
		query.Set(key, strconv.FormatInt(value, 10))
	}
}

func setTrue(query url.Values, key string, value bool) {
	if value {
		query.Set(key, "true")
	}
}

func setNonEmpty(query url.Values, key, value string) {
	if value != "" {
		query.Set(key, value)
	}
}

func setIntList[T ~int | ~int32 | ~int64](query url.Values, key string, values []T) {
	if len(values) == 0 {
		return
	}

	parts := make([]string, 0, len(values))
	for _, value := range values {
		parts = append(parts, strconv.FormatInt(int64(value), 10))
	}

	query.Set(key, strings.Join(parts, ","))
}

func setStringList(query url.Values, key string, values []string) {
	if len(values) > 0 {
		query.Set(key, strings.Join(values, ","))
	}
}
