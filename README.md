# go-jsonld

[![GoDoc](https://godoc.org/github.com/emersion/go-jsonld?status.svg)](https://godoc.org/github.com/emersion/go-jsonld)
[![Build Status](https://travis-ci.org/emersion/go-jsonld.svg?branch=master)](https://travis-ci.org/emersion/go-jsonld)

A Go library for [JSON-LD](https://www.w3.org/TR/json-ld/).

## Usage

```go
type Person struct {
	ID string `jsonld:"@id"`
	Name string `jsonld:"http://schema.org/name"`
}

const jsonString = `{
  "@id": "http://me.markus-lanthaler.com/",
  "http://schema.org/name": "Manu Sporny"
}`

var p Person
if err := jsonld.Unmarshal([]byte(jsonString), &p); err != nil {
	log.Fatal(err)
}

log.Printf(&p)
```

## License

MIT
