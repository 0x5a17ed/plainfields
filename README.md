# PlainFields 🌱

[![Go Reference](https://pkg.go.dev/badge/github.com/yourusername/plainfields.svg)](https://pkg.go.dev/github.com/yourusername/plainfields)

PlainFields is a lightweight, human-friendly configuration language for Go applications. It provides a simple syntax for expressing structured data with fields, lists, and key-value pairs.

```
name=PlainFields, version=1.0.0, ^enabled, !deprecated, features=simple;human-readable;flexible, 
settings=theme:dark;indent:4;display:compact
```


## ✨ Features

- **🔤 Clean, minimal syntax**  
  No curly braces, no complex nesting. Easy to read and write by humans
- **📏 Compact**  
  Minimal syntax overhead, no deep nesting to navigate
- **🔄 Boolean toggles**  
  Use `^feature` to enable, `!feature` to disable
- **📋 Lists and maps**  
  Simple syntax for collections: `tags=red;green;blue` `font=family:Arial;size:12`
- **⚡ Event-based parser**  
  Efficient parsing and serialization with a streaming parser API
- **🔨 Fluent builder API**  
  Programmatically create plainfields strings
- **🧩 Zero dependencies**  
  Just pure Go standard library, no `reflect` used


## 📦 Installation

```bash
go get -u github.com/0x5a17ed/plainfields
```


## 🚀 Quick Start


### 🔍 Parsing PlainFields

```go
package main

import (
	"fmt"
	
	"github.com/0x5a17ed/plainfields"
)

func main() {
	input := "^enabled, name=john, settings=theme:dark;fontSize:14"
	tokens := plainfields.Lex(input)

	for event := range plainfields.Parse(tokens) {
		switch e := event.(type) {
		case plainfields.FieldStartEvent:
			fmt.Printf("  Field: %s\n", e.Name)
		case plainfields.ValueEvent:
			fmt.Printf("  Value: %s\n", e.Value)
		case plainfields.MapStartEvent:
			fmt.Printf("    Map:\n")
		case plainfields.MapKeyEvent:
			fmt.Printf("    Key: %s\n", e.Value)
		}
	}
}

```


### ✏️ Creating PlainFields

```go
builder := plainfields.NewBuilder()
result := builder.
	Enable("feature").
	Field("name", "john").
	List("tags", "dev", "prod").
	Pairs("settings", "theme", "dark", "fontSize", 14).
	String()

fmt.Println(result)
// Output: ^feature,name=john,tags=dev;prod,settings=theme:dark;fontSize:14
```


## 📝 Syntax

PlainFields uses a simple, flat structure:

- **Fields** are comma-separated: `name=john, age=30`
- **Boolean toggles** use prefixes: `^enabled`, `!disabled`
- **Lists** use semicolons: `tags=red;green;blue`
- **Maps** (key-value pairs) use colon and semicolon: `settings=theme:dark;fontSize:14`


### 🧪 Data Types

- **🔤 Strings**: `name=john` or `message="Hello, world!"` (quotes for special chars)
- **🔢 Numbers**: `age=30`, `pi=3.14`, `hex=0xFF`, `binary=0b1010`
- **✅ Booleans**: `active=true` or `valid=false`
- **📋 Lists**: `colors=red;green;blue`
- **🗺️ Maps**: `settings=theme:dark;fontSize:14`
- **❌ Null**: `value=nil`


## 📚 Detailed Usage

### 🛠️ Builder with Options

```go
options := plainfields.BuilderOptions{
	SpaceAfterFieldSeparator:   true,
	SpaceAfterListSeparator:    true,
	SpaceAfterPairsSeparator:   true,
	SpaceAroundFieldAssignment: true,
}

builder := plainfields.NewBuilderWithOptions(options)
result := builder.
	Enable("feature").
	Field("name", "john").
	List("tags", "dev", "prod").
	Pairs("settings", "theme", "dark", "fontSize", 14).
	String()

fmt.Println(result)
// Output: ^feature, name = john, tags = dev; prod, settings = theme: dark; fontSize: 14
```


## 📜 License

This project is licensed under the 0BSD Licence — see the [LICENCE](LICENSE) file for details.

---

<p align="center">Made with ❤️ for structured data 🌟</p>
