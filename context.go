package jsonld

import (
	"errors"
	"strings"
)

type Context struct {
	Lang string
	Base string // Base URI to resolve relative URIs.
	Vocab string // Base vocabulary.
	Terms map[string]*Resource
}

func (ctx *Context) newChild(child *Context) *Context {
	if child == nil {
		child = new(Context)
	}
	if child.Terms == nil {
		child.Terms = make(map[string]*Resource)
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
