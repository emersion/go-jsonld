package jsonld

type Props map[string][]interface{}

func (p Props) Get(k string) interface{} {
	v, ok := p[k]
	if !ok || len(v) == 0 {
		return nil
	}
	return v[0]
}

func (p Props) Type() string {
	t, _ := p.Get(propType).(string)
	return t
}
