# Triple Store

Triple Store is a library to manipulate RDF triples in a fast and fluent fashion.

RDF triples allow to represent any data and its relations to other data. It is a very versatile concept and is used in [Linked Data](https://en.wikipedia.org/wiki/Linked_data), graph databases and representation, simple directed graph, etc....

Here the RDF triples implementation follows along the [W3C RDF concepts](https://www.w3.org/TR/rdf11-concepts/). (**Note that blank nodes and reification are not implemented**.). More digestible info on [RDF Wikipedia](https://en.wikipedia.org/wiki/Resource_Description_Framework)

## Triples quickstart

RDF is a resource description framework that allows to describe anything using triples. This is a powerful concept!

A triple simply consists of:

```
subject -> predicate -> object
```

... or you can also view that as: 

```
entity -> attribute -> value
```

## Library 

This library is written using the [Golang](https://golang.org) language.

Get it:

```sh
go get -u github.com/wallix/triplestore
```

Test it:

```
go test -v -cover github.com/wallix/triplestore
```

Import it in your source code:

```go
import (
	"github.com/wallix/triplestore"
	// tstore "github.com/wallix/triplestore" for less verbosity
)
```

## Usage

### Build triples

```go
triples = append(triples,
	Subject("me").Predicate("name").StringLiteral("jsmith"),
 	Subject("me").Predicate("age").IntegerLiteral(26),
 	Subject("me").Predicate("male").BooleanLiteral(true),
 	Subject("me").Predicate("born").DateTimeLiteral(time.Now()),
 	Subject("me").Predicate("mother").Resource("mum#121287"),
)
```

Although you can build triples the way you want to model any data, they are usually built from known RDF vocabularies & namespace. Ex: [foaf](http://xmlns.com/foaf/spec/), ...

### Triple Storage

// TODO

### Triple Queries

// TODO

### RDF Graph

// TODO

### Raw storage

In this library, higher level APIs used a raw `Encoder` and `Decoder` to store and exchanges triples. They are encoded & decoded using a simple format in order to persists, flush or send them over the network.

```go
triples = append(triples,
	Subject("me").Predicate("name").StringLiteral("jsmith"),
	...
 	Subject("me").Predicate("born").DateTimeLiteral(time.Now()),
)

enc := NewEncoder(myWriter)
err := enc.Encode(triples)
...

dec := NewDecoder(myReader)
triples, err := dec.Decode()

```