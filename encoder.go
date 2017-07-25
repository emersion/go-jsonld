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
	raw, err := marshalValue(reflect.ValueOf(v))
	if err != nil {
		return err
	}

	raw, err = formatValue(raw, e.Context)
	if err != nil {
		return err
	}

	if e.Context != nil {
		formattedCtx, err := formatContext(e.Context)
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

func formatValue(v interface{}, ctx *Context) (interface{}, error) {
	switch v := v.(type) {
	case *Resource:
		return formatResource(v, ctx)
	default:
		return v, nil
	}
}

func formatResource(r *Resource, ctx *Context) (map[string]interface{}, error) {
	m := make(map[string]interface{})

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
			m[k], err = formatValue(values[0], ctx)
		} else {
			m[k], err = formatValue(values, ctx)
		}
		if err != nil {
			return m, err
		}
	}

	return m, nil
}
