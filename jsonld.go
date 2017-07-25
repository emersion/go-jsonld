// Package jsonld implements JSON-LD, as defined in
// https://www.w3.org/TR/json-ld/.
package jsonld

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
)

const (
	nsRDFS = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	nsXSD = "http://www.w3.org/2001/XMLSchema#"
)

const propType = nsRDFS + "type"

const (
	typeString = nsXSD + "string"
	typeBoolean = nsXSD + "boolean"
	typeInteger = nsXSD + "integer"
	typeDouble = nsXSD + "double"
	typeAnyURI = nsXSD + "anyURI"
)

type Type struct {
	URI string
}

func formatValue(v interface{}, ctx *Context) (interface{}, error) {
	switch v := v.(type) {
	case *Resource:
		return formatResource(v, ctx)
	default:
		return v, nil
	}
}

func unmarshalValue(src interface{}, dst reflect.Value) error {
	switch src := src.(type) {
	case *Resource:
		// TODO: remove this
		return unmarshalResource(src, dst)
	default:
		rsrc := reflect.ValueOf(src)
		if dst.Type() == rsrc.Type() {
			dst.Set(rsrc)
			return nil
		} else {
			return fmt.Errorf("jsonld: cannot unmarshal %v to %v", rsrc.Type(), dst.Type())
		}
	}
}

func marshalValue(v reflect.Value) (interface{}, error) {
	switch v.Kind() {
	case reflect.Struct:
		r, err := marshalResource(v)
		if err != nil {
			return nil, err
		}
		return r, nil
	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return marshalValue(reflect.Indirect(v))
	default:
		return v.Interface(), nil
	}
}

// Unmarshal parses the JSON-LD-encoded data and stores the result in the value
// pointed to by v.
//
// Unmarshal uses the inverse of the encodings that Marshal uses, allocating
// maps, slices, and pointers as necessary, with the following additional rules:
//
// To unmarshal JSON-LD into a struct:
//
//  * If the struct has a field named JSONLDType of type Type, Unmarshal records
//    the resource type in that field.
//  * If the JSONLDType field has an associated tag of the form "type-URI", the
//    resource must have the given type or else Unmarshal returns an error.
//  * If the struct has a field whose tag is "@id", Unmarshal records the
//    resource URI in that field.
//  * If the resource has a property whose URI matches a tag formatted as
//    "property-URI", the property value is recorded in that field.
//
// To unmarshal JSON-LD into an interface value, Unmarshal uses the same rules
// as the encoding/json package, except for resources which are stored as
// *Resource.
func Unmarshal(b []byte, v interface{}) error {
	return NewDecoder(bytes.NewReader(b)).Decode(v)
}

// Marshal returns the JSON-LD encoding of v.
//
// Marshal uses the same rules as the encoding/json package, except for
// Resource values.
func Marshal(v interface{}) ([]byte, error) {
	return MarshalWithContext(v, nil)
}

// MarshalWithContext returns the JSON-LD encoding of v with the context ctx.
func MarshalWithContext(v interface{}, ctx *Context) ([]byte, error) {
	raw, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}

	formatted, err := formatValue(raw, ctx)
	if err != nil {
		return nil, err
	}

	if ctx != nil {
		formattedCtx, err := formatContext(ctx)
		if err != nil {
			return nil, err
		}
		if m, ok := formatted.(map[string]interface{}); ok {
			m["@context"] = formattedCtx
		} else {
			formatted = map[string]interface{}{
				"@context": formattedCtx,
				"@value": formatted,
			}
		}
	}

	return json.Marshal(formatted)
}
