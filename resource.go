package jsonld

import (
	"reflect"
)

type Resource struct {
	ID string
	Props Props
}

func typeField(ft reflect.StructField) (t string, ok bool) {
	if ft.Name == "JSONLDType" && ft.Type == reflect.TypeOf(Type{}) {
		return ft.Tag.Get("jsonld"), true
	}
	return "", false
}

func getFieldURI(ctx *Context, ft reflect.StructField) (uri string, ok bool) {
	k := ft.Name
	if tag := ft.Tag.Get("jsonld"); tag != "" {
		k = tag
		if k == "-" {
			return "", false
		}
	}
	if ctx != nil {
		k = ctx.expand(k)
	}
	return k, true
}
