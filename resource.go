package jsonld

import (
	"fmt"
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
