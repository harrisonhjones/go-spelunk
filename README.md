# Spelunk

Explore the depths of Go structs (and also leave your mark).

[![build status badge][travis-badge]][travis-url]
[![go doc badge][go-doc-badge]][go-doc-url]
[![go report card badge][go-reportcard-badge]][go-reportcard-url]
[![code coverage badge][codecov-badge]][codecov-url]

[travis-badge]: https://travis-ci.org/harrisonhjones/go-spelunk.svg?branch=master
[travis-url]: https://travis-ci.org/harrisonhjones/go-spelunk
[go-doc-badge]: https://godoc.org/hhj.me/go/spelunk?status.svg
[go-doc-url]: https://godoc.org/hhj.me/go/spelunk
[go-reportcard-badge]: https://goreportcard.com/badge/hhj.me/go/spelunk
[go-reportcard-url]: https://goreportcard.com/report/hhj.me/go/spelunk
[codecov-badge]: https://codecov.io/gh/harrisonhjones/go-spelunk/branch/master/graph/badge.svg
[codecov-url]: https://codecov.io/gh/harrisonhjones/go-spelunk

## Description

Spelunk recursively explores Go structs using reflection. Upon finding
a configurable tag it calls registered handle functions to inspect, or
even modify, the value. Use Spelunk to validate user input, sanitize
sensitive data before logging or outputting to a user, etc.

## Installation

*go get hhj.me/go/spelunk*

## Examples

A basic example is provided below. See the
[godoc]((http://godoc.org/hhj.me/go/spelunk)) for more examples.

```
example := struct {
    String string            `spelunk:"zero"`
    Int    int               `spelunk:"zero"`
    Float  float64           `spelunk:"zero"`
    Map    map[string]string `spelunk:"zero"`
    Slice  []string          `spelunk:"zero"`
    Array  [2]string         `spelunk:"zero"`
}{
    String: "string",
    Int:    1,
    Float:  1.2,
    Map:    map[string]string{"key": "value"},
    Slice:  []string{"foo", "bar"},
    Array:  [2]string{"baz", "qux"},
}

fmt.Printf("%+v\n", example)

s1 := spelunk.New()

s1.SetHandler("zero", spelunk.Zeroer)
s1.Spelunk(&example)

fmt.Printf("%+v\n", example)
// Output:
// {String:string Int:1 Float:1.2 Map:map[key:value] Slice:[foo bar] Array:[baz qux]}
// {String: Int:0 Float:0 Map:map[] Slice:[] Array:[ ]}
```

## Bugs / Feature Requests

- [Open a issue](https://github.com/harrisonhjones/go-spelunk/issues/new)
- Fork and [submit a pull request](https://github.com/harrisonhjones/go-spelunk/compare)

