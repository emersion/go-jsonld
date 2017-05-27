package jsonld

import (
	"encoding/json"
	"errors"
	"strings"
)

const (
	nsRDFS = "http://www.w3.org/1999/02/22-rdf-syntax-ns#"
	nsXSD = "http://www.w3.org/2001/XMLSchema#"
)

const propType = nsRDFS + "type"

const (
	typeString = nsXSD + "string"
	typeBoolean = nsXSD + "boolean"
	typeInteger = nsXSD + "integer"
	typeDouble = nsXSD + "double"
	typeAnyURI = nsXSD + "anyURI"
)

type Context struct {
	Lang string
	Base string // Base URI to resolve relative URIs.
	Vocab string // Base vocabulary.
	Terms map[string]*Node
}

func (ctx *Context) newChild(child *Context) *Context {
	if child == nil {
		child = new(Context)
	}
	if child.Terms == nil {
		child.Terms = make(map[string]*Node)
	}
	if ctx != nil {
		if child.Lang == "" {
			child.Lang = ctx.Lang
		}
		if child.Base == "" {
			child.Base = ctx.Base
		}
		if child.Vocab == "" {
			child.Vocab = ctx.Vocab
		}

		for k, v := range ctx.Terms {
			if _, ok := child.Terms[k]; !ok {
				child.Terms[k] = v
			}
		}
	}
	return child
}

func (ctx *Context) parseChild(m map[string]interface{}) (*Context, error) {
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

		v, err := parseValue(nil, v, "")
		if err != nil {
			return nil, err
		}

		var term *Node
		switch v := v.(type) {
		case *Node:
			term = v
		case string:
			term = &Node{ID: v}
		default:
			if v != nil {
				return nil, errors.New("jsonld: malformed context value")
			}
		}
		child.Terms[k] = term
	}

	return child, nil
}

func (ctx *Context) expand(u string) string {
	if ctx == nil {
		return u
	}

	if i := strings.IndexRune(u, ':'); i >= 0 {
		// It's either an absolute or a relative URI
		prefix, suffix := u[:i], u[i+1:]
		if term, ok := ctx.Terms[prefix]; ok && term != nil {
			return term.ID + suffix // Relative
		} else {
			return u // Absolute
		}
	} else {
		return ctx.Vocab + u
	}
}

type Props map[string][]interface{}

func (p Props) Get(k string) interface{} {
	v, ok := p[k]
	if !ok || len(v) == 0 {
		return ""
	}
	return v[0]
}

type Node struct {
	ID string
	Props Props
}

func parseValue(ctx *Context, v interface{}, t string) (interface{}, error) {
	m, ok := v.(map[string]interface{})
	if ok {
		v = m["@value"]
		if t == "" {
			// TODO: arrays
			t, _ = m["@type"].(string)
		}
	}

	switch t {
	case "@id":
		if m != nil {
			return parseNode(ctx, m)
		}
		if s, ok := v.(string); ok {
			return &Node{ID: s}, nil
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
		switch v := v.(type) {
		case int64:
			return v, nil
		case int:
			return int64(v), nil
		case string:
			return json.Number(v).Int64()
		default:
			return nil, errors.New("jsonld: expected an integer")
		}
	case typeBoolean:
		if b, ok := v.(bool); ok {
			return b, nil
		} else {
			return nil, errors.New("jsonld: expected a boolean")
		}
	case typeDouble:
		switch v := v.(type) {
		case float64:
			return v, nil
		case string:
			return json.Number(v).Float64()
		default:
			return nil, errors.New("jsonld: expected a double")
		}
	case typeAnyURI:
		if u, ok := v.(string); ok {
			return ctx.expand(u), nil
		} else {
			return nil, errors.New("jsonld: expected a URI")
		}
	default:
		if m != nil {
			return parseNode(ctx, m)
		} else {
			return v, nil
		}
	}
}

func parseNode(ctx *Context, m map[string]interface{}) (*Node, error) {
	n := &Node{Props: make(Props)}

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
