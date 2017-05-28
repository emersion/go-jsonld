// Package jsonld implements JSON-LD, as defined in
// https://www.w3.org/TR/json-ld/.
package jsonld

import (
	"encoding/json"
	"errors"
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

func parseValue(ctx *Context, v interface{}, t string) (interface{}, error) {
	// Type embedded in value
	m, ok := v.(map[string]interface{})
	if ok {
		v = m["@value"]
		if t == "" {
			switch rawType := m["@type"].(type) {
			case string:
				t = rawType
			case []interface{}:
				if len(rawType) > 0 {
					t, _ = rawType[0].(string)
				}
			}
		}
	}

	switch t {
	case "@id":
		if m != nil {
			return parseResource(ctx, m)
		}
		if s, ok := v.(string); ok {
			return &Resource{ID: s}, nil
		} else {
			return nil, errors.New("jsonld: expected an ID")
		}
	case typeString:
		if s, ok := v.(string); ok {
			return s, nil
		} else {
			return nil, errors.New("jsonld: expected a string")
		}
	case typeInteger:
		switch v := v.(type) {
		case int64:
			return v, nil
		case int:
			return int64(v), nil
		case string:
			return json.Number(v).Int64()
		default:
			return nil, errors.New("jsonld: expected an integer")
		}
	case typeBoolean:
		if b, ok := v.(bool); ok {
			return b, nil
		} else {
			return nil, errors.New("jsonld: expected a boolean")
		}
	case typeDouble:
		switch v := v.(type) {
		case float64:
			return v, nil
		case string:
			return json.Number(v).Float64()
		default:
			return nil, errors.New("jsonld: expected a double")
		}
	case typeAnyURI:
		if u, ok := v.(string); ok {
			return ctx.expand(u), nil
		} else {
			return nil, errors.New("jsonld: expected a URI")
		}
	default:
		if m != nil {
			return parseResource(ctx, m)
		} else {
			// No type info, return raw JSON value
			return v, nil
		}
	}
}

func formatValue(v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case *Resource:
		return formatResource(v)
	default:
		return v, nil
	}
}

func unmarshalValue(src interface{}, dst reflect.Value) error {
	switch src := src.(type) {
	case *Resource:
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
		return formatResource(r)
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
	var raw interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		return err
	}

	parsed, err := parseValue(nil, raw, "")
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("jsonld: cannot unmarshal non-pointer")
	}

	return unmarshalValue(parsed, reflect.Indirect(rv))
}

// Marshal returns the JSON-LD encoding of v.
//
// Marshal uses the smae rules as the encoding/json package, except for
// *Resource values.
func Marshal(v interface{}) ([]byte, error) {
	raw, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		return nil, err
	}

	formatted, err := formatValue(raw)
	if err != nil {
		return nil, err
	}

	return json.Marshal(formatted)
}
