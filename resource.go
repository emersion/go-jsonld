package jsonld

import (
	"errors"
	"fmt"
	"reflect"
)

type Resource struct {
	ID string
	Props Props
}

func parseResource(ctx *Context, m map[string]interface{}) (*Resource, error) {
	n := new(Resource)

	if rawCtx, ok := m["@context"]; ok {
		// TODO: case []interface{}
		var err error
		switch rawCtx := rawCtx.(type) {
		case map[string]interface{}:
			ctx, err = ctx.parseChild(rawCtx)
		case string:
			ctx, err = ctx.fetchChild(rawCtx)
		default:
			err = errors.New("jsonld: malformed context")
		}

		if err != nil {
			return n, err
		}
	}

	if rawID, ok := m["@id"]; ok {
		if id, ok := rawID.(string); ok {
			n.ID = ctx.expand(id)
		}
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
			vv, err := parseValue(ctx, v, t)
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

func typeField(ft reflect.StructField) (t string, ok bool) {
	if ft.Name == "JSONLDType" && ft.Type == reflect.TypeOf(Type{}) {
		return ft.Tag.Get("jsonld"), true
	}
	return "", false
}

func getFieldURI(ft reflect.StructField) (uri string, ok bool) {
	k := ft.Name
	if tag := ft.Tag.Get("jsonld"); tag != "" {
		k = tag
		if k == "-" {
			return "", false
		}
	}
	return k, true
}

func unmarshalResource(r *Resource, v reflect.Value) error {
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
			k, ok := getFieldURI(ft)
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
				if err := unmarshalValue(fv, f); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func marshalResource(v reflect.Value) (*Resource, error) {
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
			k, ok := getFieldURI(ft)
			if !ok {
				continue
			}

			if k == "@id" {
				r.ID = f.String()
			} else {
				raw, err := marshalValue(f)
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
