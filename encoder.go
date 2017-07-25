package jsonld

import (
	"encoding/json"
	"io"
	"reflect"
)

// Encoder encodes JSON-LD values.
type Encoder struct {
	// If specified, this context will be used when encoding values.
	Context *Context

	enc *json.Encoder
}

// NewEncoder creates a new JSON-LD encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{enc: json.NewEncoder(w)}
}

// Encode encodes a JSON-LD value.
func (e *Encoder) Encode(v interface{}) error {
	raw, err := e.marshal(reflect.ValueOf(v))
	if err != nil {
		return err
	}

	raw, err = e.format(raw)
	if err != nil {
		return err
	}

	if e.Context != nil {
		formattedCtx, err := e.formatContext(e.Context)
		if err != nil {
			return err
		}
		if m, ok := raw.(map[string]interface{}); ok {
			m["@context"] = formattedCtx
		} else {
			raw = map[string]interface{}{
				"@context": formattedCtx,
				"@value": raw,
			}
		}
	}

	return e.enc.Encode(raw)
}

func (e *Encoder) format(v interface{}) (interface{}, error) {
	switch v := v.(type) {
	case *Resource:
		return e.formatResource(v)
	default:
		return v, nil
	}
}

func (e *Encoder) formatResource(r *Resource) (map[string]interface{}, error) {
	m := make(map[string]interface{})
	ctx := e.Context

	if r.ID != "" {
		// TODO: use ctx.Base to produce relative URIs when possible
		m["@id"] = r.ID
	}

	for k, values := range r.Props {
		if k == propType {
			k = "@type"

			for i, v := range values {
				if s, ok := v.(string); ok {
					values[i], _ = ctx.reduce(s)
				}
			}
		}

		k, term := ctx.reduce(k)
		if term != nil && term.Props.hasType("@id") {
			for i, v := range values {
				if r, ok := v.(*Resource); ok && len(r.Props) == 0 {
					values[i] = r.ID
				}
			}
		}

		var err error
		if len(values) == 1 {
			m[k], err = e.format(values[0])
		} else {
			m[k], err = e.format(values)
		}
		if err != nil {
			return m, err
		}
	}

	return m, nil
}

func (e *Encoder) marshal(v reflect.Value) (interface{}, error) {
	switch v.Kind() {
	case reflect.Struct:
		r, err := e.marshalResource(v)
		if err != nil {
			return nil, err
		}
		return r, nil
	case reflect.Ptr:
		if v.IsNil() {
			return nil, nil
		}
		return e.marshal(reflect.Indirect(v))
	default:
		return v.Interface(), nil
	}
}

func (e *Encoder) marshalResource(v reflect.Value) (*Resource, error) {
	// TODO: don't panic

	// TODO: use &Resource instead
	if v.Type() == reflect.TypeOf(Resource{}) {
		r := v.Interface().(Resource)
		return &r, nil
	}

	r := new(Resource)

	t := v.Type()

	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)

		if typeURI, ok := typeField(ft); ok {
			r.Props.Set(propType, typeURI)
		} else {
			k, ok := getFieldURI(e.Context, ft)
			if !ok {
				continue
			}

			if k == "@id" {
				r.ID = f.String()
			} else {
				raw, err := e.marshal(f)
				if err != nil {
					return r, err
				}

				if r.Props == nil {
					r.Props = make(Props)
				}

				r.Props.Add(k, raw)
			}
		}
	}

	return r, nil
}

func (e *Encoder) formatContext(ctx *Context) (interface{}, error) {
	if ctx == nil {
		return nil, nil
	}
	if ctx.URL != "" {
		return ctx.URL, nil
	}

	m := make(map[string]interface{})

	if ctx.Lang != "" {
		m["@lang"] = ctx.Lang
	}
	if ctx.Base != "" {
		m["@base"] = ctx.Base
	}
	if ctx.Vocab != "" {
		m["@vocab"] = ctx.Vocab
	}

	for k, term := range ctx.Terms {
		if len(term.Props) == 0 {
			m[k] = term.ID
		} else {
			raw, err := e.formatResource(term)
			if err != nil {
				return m, err
			}
			m[k] = raw
		}
	}

	return m, nil
}
