package jsonld

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
)

// FetchContextFunc fetches remote contexts.
type FetchContextFunc func(url string) (*Context, error)

// FetchContext fetches remote contexts with http.DefaultClient.
func FetchContext(url string) (*Context, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Accept", "application/ld+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// TODO: that's ugly, I'm just lazy
	var data struct {
		Context interface{} `json:"@context"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return new(Decoder).parseContext(nil, data.Context)
}

// Decoder decodes JSON-LD values.
type Decoder struct {
	// Context, if non-nil, will be used when decoding values.
	Context *Context
	// FetchContext, if non-nil, will be called to fetch remote contexts. By
	// default, remote contexts are not fetched.
	FetchContext FetchContextFunc

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

	return d.unmarshal(raw, reflect.Indirect(rv))
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
		var err error
		if ctx, err = d.parseContext(ctx, rawCtx); err != nil {
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

func (d *Decoder) parseContext(ctx *Context, v interface{}) (*Context, error) {
	var err error
	switch v := v.(type) {
	case []interface{}:
		for _, vv := range v {
			ctx, err = d.parseContext(ctx, vv)
			if err != nil {
				return nil, err
			}
		}
	case map[string]interface{}:
		ctx, err = d.parseContextMap(ctx, v)
	case string:
		ctx, err = d.fetchContext(ctx, v)
	default:
		err = errors.New("jsonld: malformed context")
	}
	return ctx, err
}

func (d *Decoder) parseContextMap(ctx *Context, m map[string]interface{}) (*Context, error) {
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
	// TODO: inherit from ctx
	if d.FetchContext != nil {
		return d.FetchContext(url)
	}
	return nil, errors.New("jsonld: fetching remote contexts is disabled")
}

func (d *Decoder) unmarshal(src interface{}, dst reflect.Value) error {
	switch src := src.(type) {
	case *Resource:
		return d.unmarshalResource(src, dst)
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

func (d *Decoder) unmarshalResource(r *Resource, v reflect.Value) error {
	// TODO: do not panic

	t := v.Type()

	if t == reflect.TypeOf(Resource{}) {
		// TODO: do not copy value
		v.Set(reflect.ValueOf(*r))
		return nil
	}

	for i := 0; i < t.NumField(); i++ {
		f := v.Field(i)
		ft := t.Field(i)

		if wantTypeURI, ok := typeField(ft); ok {
			typeURI := r.Props.Type()
			if wantTypeURI != typeURI {
				return fmt.Errorf("jsonld: mismatched type %v", wantTypeURI)
			}
			f.Set(reflect.ValueOf(Type{typeURI}))
		} else {
			k, ok := getFieldURI(d.Context, ft)
			if !ok {
				continue
			}

			if k == "@id" {
				f.SetString(r.ID)
			} else {
				fv := r.Props.Get(k)
				if fv == nil {
					continue
				}

				if f.Kind() == reflect.Ptr {
					if f.IsNil() {
						f.Set(reflect.New(f.Type().Elem()))
					}
					f = reflect.Indirect(f)
				}
				if err := d.unmarshal(fv, f); err != nil {
					return err
				}
			}
		}
	}

	return nil
}
