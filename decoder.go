package jsonld

import (
	"encoding/json"
	"errors"
	"io"
	"reflect"
)

// Decoder decodes JSON-LD values.
type Decoder struct {
	dec *json.Decoder
}

// NewDecoder creates a new JSON-LD decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{dec: json.NewDecoder(r)}
}

// Decode decodes a JSON-LD value.
func (d *Decoder) Decode(v interface{}) error {
	var raw interface{}
	if err := d.dec.Decode(&raw); err != nil {
		return err
	}

	raw, err := d.parse(nil, raw, "")
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return errors.New("jsonld: cannot unmarshal non-pointer")
	}

	return unmarshalValue(raw, reflect.Indirect(rv))
}

func (d *Decoder) parse(ctx *Context, v interface{}, t string) (interface{}, error) {
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
			return d.parseResource(ctx, m)
		} else if s, ok := v.(string); ok {
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
		// TODO: big ints are strings, use json.Number(v).Int64()
		if f, ok := v.(float64); ok {
			return int64(f), nil
		} else {
			return nil, errors.New("jsonld: expected an integer")
		}
	case typeBoolean:
		if b, ok := v.(bool); ok {
			return b, nil
		} else {
			return nil, errors.New("jsonld: expected a boolean")
		}
	case typeDouble:
		// TODO: big floats are strings, use json.Number(v).Float64()
		if f, ok := v.(float64); ok {
			return f, nil
		} else {
			return nil, errors.New("jsonld: expected an double")
		}
	case typeAnyURI:
		if u, ok := v.(string); ok {
			return ctx.expand(u), nil
		} else {
			return nil, errors.New("jsonld: expected a URI")
		}
	default:
		if m != nil {
			return d.parseResource(ctx, m)
		} else {
			// No type info, return raw JSON value
			return v, nil
		}
	}
}

func (d *Decoder) parseResource(ctx *Context, m map[string]interface{}) (*Resource, error) {
	if rawCtx, ok := m["@context"]; ok {
		// TODO: case []interface{}
		var err error
		switch rawCtx := rawCtx.(type) {
		case map[string]interface{}:
			ctx, err = d.parseContext(ctx, rawCtx)
		case string:
			ctx, err = d.fetchContext(ctx, rawCtx)
		default:
			err = errors.New("jsonld: malformed context")
		}

		if err != nil {
			return nil, err
		}
	}

	n := new(Resource)

	if rawID, ok := m["@id"].(string); ok {
		n.ID = ctx.expand(rawID)
	}

	for k, v := range m {
		if k == "@type" {
			k = propType
		}

		if len(k) > 0 && k[0] == '@' {
			continue
		}

		var t string
		if k == propType {
			t = typeAnyURI
		} else if ctx != nil {
			if term, ok := ctx.Terms[k]; ok {
				if term != nil {
					if term.ID != "" {
						k = ctx.expand(term.ID)
					} else {
						k = ctx.expand(k)
					}
					t, _ = term.Props.Get(propType).(string)
				} else {
					// Keep k as-is
				}
			} else {
				k = ctx.expand(k)
			}
		}

		values, ok := v.([]interface{})
		if !ok {
			values = []interface{}{v}
		}

		for _, v := range values {
			vv, err := d.parse(ctx, v, t)
			if err != nil {
				return n, err
			}
			if n.Props == nil {
				n.Props = make(Props)
			}
			n.Props[k] = append(n.Props[k], vv)
		}
	}

	return n, nil
}

func (d *Decoder) parseContext(ctx *Context, m map[string]interface{}) (*Context, error) {
	child := ctx.newChild(nil)

	if lang, ok := m["@lang"].(string); ok {
		child.Lang = lang
	}
	if base, ok := m["@base"].(string); ok {
		child.Base = base
	}
	if vocab, ok := m["@vocab"].(string); ok {
		child.Vocab = vocab
	}

	for k, v := range m {
		if len(k) > 0 && k[0] == '@' {
			continue
		}

		v, err := d.parse(nil, v, "")
		if err != nil {
			return nil, err
		}

		var term *Resource
		switch v := v.(type) {
		case *Resource:
			term = v
		case string:
			term = &Resource{ID: v}
		default:
			if v != nil {
				return nil, errors.New("jsonld: malformed context value")
			}
		}
		child.Terms[k] = term
	}

	return child, nil
}

func (d *Decoder) fetchContext(ctx *Context, url string) (*Context, error) {
	return nil, errors.New("jsonld: fetching remote contexts is not yet implemented")
}
