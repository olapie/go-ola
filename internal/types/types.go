package types

import "reflect"

var MapStringToStringType = reflect.TypeOf(map[string]string(nil))
var MapStringToStringSliceType = reflect.TypeOf(map[string][]string(nil))
