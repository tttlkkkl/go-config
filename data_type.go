package conf

import "reflect"

func getTypeOf(v interface{}) string {
	return reflect.TypeOf(v).String()
}
