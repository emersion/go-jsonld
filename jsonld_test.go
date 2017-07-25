package jsonld

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

type person struct {
	ID string `jsonld:"@id"`
	Name string `jsonld:"http://schema.org/name"`
	URL *Resource `jsonld:"http://schema.org/url"`
	Image *Resource `jsonld:"http://schema.org/image"`
}

var personContext = &Context{
	URL: "http://json-ld.org/contexts/person.jsonld",
	Terms: map[string]*Resource{
		"name": {ID: "http://schema.org/name"},
		"image": {
			ID: "http://schema.org/image",
			Props: Props{propType: {"@id"}},
		},
		"homepage": {
			ID: "http://schema.org/url",
			Props: Props{propType: {"@id"}},
		},
	},
}

type personWithContext struct {
	ID string `jsonld:"@id"`
	Name string `jsonld:"name"`
	URL *Resource `jsonld:"homepage"`
	Image *Resource `jsonld:"image"`
}

type restaurant struct {
	ID string `jsonld:"@id"`
	Name string `jsonld:"http://schema.org/name"`
	DatabaseID string `jsonld:"databaseId"`
}

type foafPerson struct {
	JSONLDType Type `jsonld:"http://xmlns.com/foaf/0.1/Person"`
	ID string `jsonld:"@id"`
	Name string `jsonld:"http://xmlns.com/foaf/0.1/name"`
	Homepage *Resource `jsonld:"http://xmlns.com/foaf/0.1/homepage"`
	Depiction *Resource `jsonld:"http://xmlns.com/foaf/0.1/depiction"`
}

const example2 = `{
  "http://schema.org/name": "Manu Sporny",
  "http://schema.org/url": { "@id": "http://manu.sporny.org/" },
  "http://schema.org/image": { "@id": "http://manu.sporny.org/images/manu.png" }
}`

var example2Out = &person{
	Name: "Manu Sporny",
	URL: &Resource{ID: "http://manu.sporny.org/"},
	Image: &Resource{ID: "http://manu.sporny.org/images/manu.png"},
}

var example2OutWithContext = &personWithContext{
	Name: "Manu Sporny",
	URL: &Resource{ID: "http://manu.sporny.org/"},
	Image: &Resource{ID: "http://manu.sporny.org/images/manu.png"},
}

var example2Resource = &Resource{
	Props: Props{
		"http://schema.org/name": {"Manu Sporny"},
		"http://schema.org/url": {&Resource{ID: "http://manu.sporny.org/"}},
		"http://schema.org/image": {&Resource{ID: "http://manu.sporny.org/images/manu.png"}},
	},
}

const example4 = `{
  "@context": "http://json-ld.org/contexts/person.jsonld",
  "name": "Manu Sporny",
  "homepage": "http://manu.sporny.org/",
  "image": "http://manu.sporny.org/images/manu.png"
}`

var example4Out = example2Out

const example5 = `{
  "@context":
  {
    "name": "http://schema.org/name",
    "image": {
      "@id": "http://schema.org/image",
      "@type": "@id"
    },
    "homepage": {
      "@id": "http://schema.org/url",
      "@type": "@id"
    }
  },
  "name": "Manu Sporny",
  "homepage": "http://manu.sporny.org/",
  "image": "http://manu.sporny.org/images/manu.png"
}`

var example5Out = example2Out

// TODO
const example9 = `{
  "@context":
  {
    "name": "http://schema.org/name"
  },
  "name": "Manu Sporny",
  "status": "trollin'"
}`

type personWithStatus struct {
	person
	Status string `jsonld:"status"`
}

var example9Out = &personWithStatus{
	person: person{
		Name: "Manu Sporny",
	},
	Status: "trollin'",
}

const example11 = `{
  "@context":
  {
    "name": "http://schema.org/name"
  },
  "@id": "http://me.markus-lanthaler.com/",
  "name": "Markus Lanthaler"
}`

var example11Out = &person{
	ID: "http://me.markus-lanthaler.com/",
	Name: "Markus Lanthaler",
}

const example12 = `{
  "@id": "http://example.org/places#BrewEats",
  "@type": "http://schema.org/Restaurant"
}`

var example12Out = &restaurant{
	ID: "http://example.org/places#BrewEats",
}

// TODO
const example13 = `{
  "@id": "http://example.org/places#BrewEats",
  "@type": [ "http://schema.org/Restaurant", "http://schema.org/Brewery" ]
}`

const example17 = `{
  "@context": {
    "@vocab": "http://schema.org/"
  },
  "@id": "http://example.org/places#BrewEats",
  "@type": "Restaurant",
  "name": "Brew Eats"
}`

var example17Out = &restaurant{
	ID: "http://example.org/places#BrewEats",
	Name: "Brew Eats",
}

const example18 = `{
  "@context":
  {
     "@vocab": "http://schema.org/",
     "databaseId": null
  },
  "@id": "http://example.org/places#BrewEats",
  "@type": "Restaurant",
  "name": "Brew Eats",
  "databaseId": "23987520"
}`

var example18Out = &restaurant{
	ID: "http://example.org/places#BrewEats",
	Name: "Brew Eats",
	DatabaseID: "23987520",
}

const example19 = `{
  "@context":
  {
    "foaf": "http://xmlns.com/foaf/0.1/"
  },
  "@type": "foaf:Person",
  "foaf:name": "Dave Longley"
}`

var example19Out = &foafPerson{
	JSONLDType: Type{URI: "http://xmlns.com/foaf/0.1/Person"},
	Name: "Dave Longley",
}

const example20 = `{
  "@context":
  {
    "xsd": "http://www.w3.org/2001/XMLSchema#",
    "foaf": "http://xmlns.com/foaf/0.1/",
    "foaf:homepage": { "@type": "@id" },
    "picture": { "@id": "foaf:depiction", "@type": "@id" }
  },
  "@id": "http://me.markus-lanthaler.com/",
  "@type": "foaf:Person",
  "foaf:name": "Markus Lanthaler",
  "foaf:homepage": "http://www.markus-lanthaler.com/",
  "picture": "http://twitter.com/account/profile_image/markuslanthaler"
}`

const example20MultipleContexts = `{
  "@context":
  [
    {
      "foaf": "http://xmlns.com/foaf/0.1/",
      "foaf:homepage": { "@type": "@id" }
    },
    {
      "picture": { "@id": "foaf:depiction", "@type": "@id" }
    }
  ],
  "@id": "http://me.markus-lanthaler.com/",
  "@type": "foaf:Person",
  "foaf:name": "Markus Lanthaler",
  "foaf:homepage": "http://www.markus-lanthaler.com/",
  "picture": "http://twitter.com/account/profile_image/markuslanthaler"
}`

var example20Out = &foafPerson{
	JSONLDType: Type{URI: "http://xmlns.com/foaf/0.1/Person"},
	ID: "http://me.markus-lanthaler.com/",
	Name: "Markus Lanthaler",
	Homepage: &Resource{ID: "http://www.markus-lanthaler.com/"},
	Depiction: &Resource{ID: "http://twitter.com/account/profile_image/markuslanthaler"},
}

var unmarshalTests = []struct{
	jsonld string
	in interface{}
	out interface{}
	ctx *Context
	fetch FetchContextFunc
}{
	{
		jsonld: example2,
		in: &Resource{},
		out: example2Resource,
	},
	{
		jsonld: example2,
		in: &person{},
		out: example2Out,
	},
	{
		jsonld: example4,
		in: &person{},
		out: example4Out,
		fetch: func(url string) (*Context, error) {
			if url != "http://json-ld.org/contexts/person.jsonld" {
				return nil, fmt.Errorf("invalid context URL: %v", url)
			}
			return personContext, nil
		},
	},
	{
		jsonld: example5,
		in: &person{},
		out: example5Out,
	},
	{
		jsonld: example11,
		in: &person{},
		out: example11Out,
	},
	{
		jsonld: example12,
		in: &restaurant{},
		out: example12Out,
	},
	{
		jsonld: example17,
		in: &restaurant{},
		out: example17Out,
	},
	{
		jsonld: example18,
		in: &restaurant{},
		out: example18Out,
	},
	{
		jsonld: example19,
		in: &foafPerson{},
		out: example19Out,
	},
	{
		jsonld: example20,
		in: &foafPerson{},
		out: example20Out,
	},
	{
		jsonld: example20MultipleContexts,
		in: &foafPerson{},
		out: example20Out,
	},
	{
		jsonld: example2,
		in: &personWithContext{},
		ctx: personContext,
		out: example2OutWithContext,
	},
}

func TestUnmarshal(t *testing.T) {
	for _, test := range unmarshalTests {
		dec := NewDecoder(strings.NewReader(test.jsonld))
		dec.Context = test.ctx
		dec.FetchContext = test.fetch

		v := test.in
		if err := dec.Decode(v); err != nil {
			t.Errorf("unmarshalResource(%v) = %v", test.jsonld, err)
		} else if !reflect.DeepEqual(test.out, v) {
			t.Errorf("unmarshalResource(%v) = %#v, want %#v", test.jsonld, v, test.out)
		}
	}
}

var marshalTests = []struct{
	jsonld string
	in interface{}
	ctx *Context
}{
	{
		jsonld: example2,
		in: example2Resource,
	},
	{
		jsonld: example2,
		in: example2Out,
	},
	{
		jsonld: example4,
		in: example4Out,
		ctx: &Context{
			URL: "http://json-ld.org/contexts/person.jsonld",
			Terms: map[string]*Resource{
				"name": {ID: "http://schema.org/name"},
				"image": {
					ID: "http://schema.org/image",
					Props: Props{propType: {"@id"}},
				},
				"homepage": {
					ID: "http://schema.org/url",
					Props: Props{propType: {"@id"}},
				},
			},
		},
	},
	{
		jsonld: example5,
		in: example5Out,
		ctx: &Context{
			Terms: map[string]*Resource{
				"name": {ID: "http://schema.org/name"},
				"image": {
					ID: "http://schema.org/image",
					Props: Props{propType: {"@id"}},
				},
				"homepage": {
					ID: "http://schema.org/url",
					Props: Props{propType: {"@id"}},
				},
			},
		},
	},
}

func TestMarshalWithContext(t *testing.T) {
	for _, test := range marshalTests {
		var want interface{}
		if err := json.Unmarshal([]byte(test.jsonld), &want); err != nil {
			t.Fatalf("json.Unmarshal(want = %v) = %v", test.jsonld, err)
		}

		b, err := MarshalWithContext(test.in, test.ctx)
		if err != nil {
			t.Errorf("MarshalWithContext(%#v, %#v) = %v", test.in, test.ctx, err)
		} else {
			var v interface{}
			if err := json.Unmarshal(b, &v); err != nil {
				t.Fatalf("json.Unmarshal(got = %v) = %v", string(b), err)
			}

			if !reflect.DeepEqual(v, want) {
				t.Errorf("MarshalWithContext(%#v, %#v) = %v, want %v", test.in, test.ctx, string(b), test.jsonld)
			}
		}
	}
}
