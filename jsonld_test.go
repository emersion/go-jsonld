package jsonld

import (
	"encoding/json"
	"reflect"
	"testing"
)

type person struct {
	ID string `jsonld:"@id"`
	Name string `jsonld:"http://schema.org/name"`
	URL *Resource `jsonld:"http://schema.org/url"`
	Image *Resource `jsonld:"http://schema.org/image"`
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

// TODO
const example4 = `{
  "@context": "http://json-ld.org/contexts/person.jsonld",
  "name": "Manu Sporny",
  "homepage": "http://manu.sporny.org/",
  "image": "http://manu.sporny.org/images/manu.png"
}`

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
}{
	{
		jsonld: example2,
		in: &person{},
		out: example2Out,
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
}

func TestUnmarshal(t *testing.T) {
	for _, test := range unmarshalTests {
		v := test.in
		if err := Unmarshal([]byte(test.jsonld), v); err != nil {
			t.Errorf("unmarshalResource(%v) = %v", test.jsonld, err)
		} else if !reflect.DeepEqual(test.out, v) {
			t.Errorf("unmarshalResource(%v) = %#v, want %#v", test.jsonld, v, test.out)
		}
	}
}

var marshalTests = []struct{
	jsonld string
	in interface{}
}{
	{
		jsonld: example2,
		in: example2Out,
	},
}

func TestMarshal(t *testing.T) {
	for _, test := range marshalTests {
		var want interface{}
		if err := json.Unmarshal([]byte(test.jsonld), &want); err != nil {
			t.Fatalf("json.Unmarshal(want = %v) = %v", test.jsonld, err)
		}

		b, err := Marshal(test.in)
		if err != nil {
			t.Errorf("Marshal(%#v) = %v", test.in, err)
		} else {
			var v interface{}
			if err := json.Unmarshal(b, &v); err != nil {
				t.Fatalf("json.Unmarshal(got = %v) = %v", string(b), err)
			}

			if !reflect.DeepEqual(v, want) {
				t.Errorf("Marshal(%#v) = %v, want %v", test.in, string(b), test.jsonld)
			}
		}
	}
}
