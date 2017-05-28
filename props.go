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

func (p Props) hasType(t string) bool {
	for _, v := range p[propType] {
		s, _ := v.(string)
		if s == t {
			return true
		}
	}
	return false
}

func (p Props) Set(k string, v interface{}) {
	p[k] = []interface{}{v}
}

func (p Props) Add(k string, v interface{}) {
	p[k] = append(p[k], v)
}
