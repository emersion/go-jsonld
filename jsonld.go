package jsonld

import (
	"encoding/json"
	"errors"
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

func unmarshalValue(src interface{}, dst reflect.Value) error {
	// TODO: do not panic
	switch src := src.(type) {
	case *Resource:
		return unmarshalResource(src, dst)
	default:
		dst.Set(reflect.ValueOf(src))
		return nil
	}
}

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
