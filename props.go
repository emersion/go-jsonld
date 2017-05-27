package jsonld

type Props map[string][]interface{}

func (p Props) Get(k string) interface{} {
	v, ok := p[k]
	if !ok || len(v) == 0 {
		return ""
	}
	return v[0]
}
