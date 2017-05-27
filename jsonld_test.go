package jsonld

import (
	"encoding/json"
	"testing"
)

const example2 = `{
  "http://schema.org/name": "Manu Sporny",
  "http://schema.org/url": { "@id": "http://manu.sporny.org/" },
  "http://schema.org/image": { "@id": "http://manu.sporny.org/images/manu.png" }
}`

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

const example9 = `{
  "@context":
  {
    "name": "http://schema.org/name"
  },
  "name": "Manu Sporny",
  "status": "trollin'"
}`

const example11 = `{
  "@context":
  {
    "name": "http://schema.org/name"
  },
  "@id": "http://me.markus-lanthaler.com/",
  "name": "Markus Lanthaler"
}`

const example12 = `{
  "@id": "http://example.org/places#BrewEats",
  "@type": "http://schema.org/Restaurant"
}`

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

const example19 = `{
  "@context":
  {
    "foaf": "http://xmlns.com/foaf/0.1/"
  },
  "@type": "foaf:Person",
  "foaf:name": "Dave Longley"
}`

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

func TestParseValue(t *testing.T) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(example13), &m); err != nil {
		t.Fatalf("json.Unmarshal() = %v", err)
	}

	v, err := parseValue(nil, m, "")
	if err != nil {
		t.Fatalf("parseValue() = %v", err)
	}

	t.Log(v)
}

type person struct {
	ID string `jsonld:"@id"`
	Name string `jsonld:"http://xmlns.com/foaf/0.1/name"`
	Homepage string `jsonld:"http://xmlns.com/foaf/0.1/homepage"`
	Depiction string `jsonld:"http://xmlns.com/foaf/0.1/depiction"`
}

func TestUnmarshal(t *testing.T) {
	var p person
	if err := Unmarshal([]byte(example20), &p); err != nil {
		t.Fatalf("unmarshalResource() = %v", err)
	}

	t.Logf("%#v", &p)
}
