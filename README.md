# go-jsonld

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
