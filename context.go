package jsonld

import (
	"strings"
)

type Context struct {
	URL string
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
		if term, ok := ctx.Terms[u]; ok && term != nil {
			return term.ID
		} else {
			return ctx.Vocab + u
		}
	}
}

func (ctx *Context) reduce(u string) (reduced string, term *Resource) {
	if ctx == nil {
		return u, nil
	}

	for k, term := range ctx.Terms {
		if len(k) > 0 && k[0] == '@' {
			continue
		}
		if term.ID == u {
			return k, term
		}
		if term.ID != "" && strings.HasPrefix(u, term.ID) {
			return k + ":" + strings.TrimPrefix(u, term.ID), nil
		}
	}

	if ctx.Vocab != "" && strings.HasPrefix(u, ctx.Vocab) {
		return strings.TrimPrefix(u, ctx.Vocab), nil
	}

	return u, nil
}
