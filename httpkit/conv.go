package httpkit

import (
	"net/http"
	"reflect"
	"strings"
)

type MapOutput interface {
	map[string]string | map[string]any
}

type MapInput interface {
	http.Header | []*http.Cookie
}

func ToMap[IN MapInput, OUT MapOutput](input IN) OUT {
	var out OUT
	var res any
	_, isMapString := any(out).(map[string]string)
	switch v := any(input).(type) {
	case http.Header:
		if isMapString {
			res = headerToMapString(v)
		} else {
			res = headerToMap(v)
		}
	case []*http.Cookie:
		if isMapString {
			res = cookiesToMapString(v)
		} else {
			res = cookiesToMap(v)
		}
	}

	if res != nil {
		reflect.ValueOf(&out).Elem().Set(reflect.ValueOf(res))
	}
	return out
}

func ToMapAny[T MapInput](v T) map[string]any {
	return ToMap[T, map[string]any](v)
}

func ToMapString[T MapInput](v T) map[string]string {
	return ToMap[T, map[string]string](v)
}

func headerToMap(h http.Header) map[string]any {
	m := make(map[string]any, len(h))
	for k, v := range h {
		if len(v) == 0 {
			continue
		}

		if len(v) == 1 {
			m[k] = v[0]
			continue
		}
		m[k] = v
		k = strings.ToLower(k)
		if strings.HasPrefix(k, "x-") {
			k = k[2:]
			k = strings.Replace(k, "-", "_", -1)
			if _, ok := m[k]; !ok {
				m[k] = v
			}
		}
	}
	return m
}

func headerToMapString(h http.Header) map[string]string {
	m := make(map[string]string, len(h))
	for k, v := range h {
		if len(v) == 0 {
			continue
		}
		m[k] = v[0]
		k = strings.ToLower(k)
		if strings.HasPrefix(k, "x-") {
			k = k[2:]
			k = strings.Replace(k, "-", "_", -1)
			if _, ok := m[k]; !ok {
				m[k] = v[0]
			}
		}
	}
	return m
}

func cookiesToMapString(cookies []*http.Cookie) map[string]string {
	m := make(map[string]string, len(cookies))
	for _, c := range cookies {
		m[c.Name] = c.Value
	}
	return m
}

func cookiesToMap(cookies []*http.Cookie) map[string]any {
	m := make(map[string]any, len(cookies))
	for _, c := range cookies {
		m[c.Name] = c.Value
	}
	return m
}
