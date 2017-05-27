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
	n := &Resource{Props: make(Props)}

	if rawCtx, ok := m["@context"]; ok {
		// TODO: string, array
		switch rawCtx := rawCtx.(type) {
		case map[string]interface{}:
			var err error
			ctx, err = ctx.parseChild(rawCtx)
			if err != nil {
				return n, err
			}
		default:
			return n, errors.New("jsonld: malformed context")
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
			n.Props[k] = append(n.Props[k], vv)
		}
	}

	return n, nil
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

		k := ft.Name
		if tag := ft.Tag.Get("jsonld"); tag != "" {
			k = tag
		}

		if ft.Name == "JSONLDType" && f.Type() == reflect.TypeOf(Type{}) {
			typeURI := r.Props.Type()
			if tag := ft.Tag.Get("jsonld"); tag != "" {
				if tag != typeURI {
					return fmt.Errorf("jsonld: mismatched type %v", tag)
				}
			}
			f.Set(reflect.ValueOf(Type{typeURI}))
		} else if k == "@id" {
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

	return nil
}
