# Kaval 🌱

[![Go Reference](https://pkg.go.dev/badge/github.com/0x5a17ed/kaval.svg)](https://pkg.go.dev/github.com/0x5a17ed/kaval)
[![License: 0BSD](https://img.shields.io/badge/License-0BSD-blue.svg)](https://opensource.org/licenses/0BSD)

Kaval is a lightweight, human-friendly configuration language for Go applications. It provides a simple syntax for expressing structured data with fields, lists, and key-value pairs.

```
Kaval, version=1.0.0, ^enabled, !deprecated, features=simple;human-readable;flexible, 
settings=theme:dark;indent:4;display:compact
```


## ✨ Features

- **🔤 Clean, minimal syntax**  
  No curly braces, no complex nesting. Friendly to read and write by humans
- **📏 Compact**  
  Minimal syntax overhead, no deep nesting to navigate
- **🔄 Convenient boolean toggles**  
  Use `^feature` to enable, `!feature` to disable
- **📋 Lists and maps**  
  Simple syntax for collections: `tags=red;green;blue` `font=family:Arial;size:12`
- **⚡ Event-based parser**  
  Efficient parsing and serialization with a streaming parser API
- **🔨 Fluent builder API**  
  Programmatically create kaval strings
- **🧩 Zero dependencies**  
  Just pure Go standard library, no `reflect` used


## 📦 Installation

```bash
go get -u github.com/0x5a17ed/kaval
```


## 🚀 Quick Start


### 🔍 Parsing Kaval

```go
package main

import (
	"fmt"
	
	"github.com/0x5a17ed/kaval"
)

func main() {
	input := "^enabled, name=john, settings=theme:dark;fontSize:14"
	for event := range kaval.Parse(input) {
		switch e := event.(type) {
		case kaval.ListStartEvent:
			fmt.Printf(" ListStart:\n")
		case kaval.ListEndEvent:
			fmt.Printf("   ListEnd:\n")
		case kaval.MapStartEvent:
			fmt.Printf("  MapStart:\n")
		case kaval.MapEndEvent:
			fmt.Printf("    MapEnd:\n")
		case kaval.ValueEvent:
			fmt.Printf("     Value: %s\n", e.Value)
		case kaval.MapKeyEvent:
			fmt.Printf("       Key: %s\n", e.Value)
		}
	}
}

// Output:
//  MapStart:
//       Key: enabled (identifier)
//     Value: true (boolean)
//       Key: name (identifier)
//     Value: john (identifier)
//       Key: settings (identifier)
//  MapStart:
//       Key: theme (identifier)
//     Value: dark (identifier)
//       Key: fontSize (identifier)
//     Value: 14 (number)
//    MapEnd:
//    MapEnd:
```


### ✏️ Creating Kaval

```go
package main

import (
	"fmt"

	"github.com/0x5a17ed/kaval"
)

func main() {
	builder := kaval.NewBuilder()
	result := builder.
		Enable("feature").
		Labeled("name", "john").
		LabeledList("tags", "dev", "prod").
		LabeledDict("settings", "theme", "dark", "fontSize", 14).
		String()

	fmt.Println(result)
}

// Output: ^feature,name=john,tags=dev;prod,settings=theme:dark;fontSize:14
```


## 📝 Syntax

Kaval uses a simple, flat structure:

- **Fields** are comma-separated: `name=john, age=30`
- **Boolean toggles** use prefixes: `^enabled`, `!disabled`
- **Lists** use semicolons: `tags=red;green;blue`
- **Maps** (key-value pairs) use colon and semicolon: `settings=theme:dark;fontSize:14`
- **Blank spaces** are allowed around syntax elements


### 🧪 Data Types

- **Strings**: `name=john` or `message="Hello, world!"` (quotes for special chars)
- **Numbers**: `age=30`, `pi=3.14`, `hex=0xFF`, `binary=0b1010`
- **Booleans**: `active=true` or `valid=false`
- **Lists**: `colors=red;green;blue`
- **Maps**: `settings=theme:dark;fontSize:14`
- **Null**: `value=nil`


## 📚 Detailed Usage

### 🛠️ Builder with Options

```go
package main

import (
	"fmt"

	"github.com/0x5a17ed/kaval"
)

func main() {
	options := kaval.BuilderOptions{
		SpaceAfterFieldSeparator:   true,
		SpaceAfterListSeparator:    true,
		SpaceAfterPairsSeparator:   true,
		SpaceAroundFieldAssignment: true,
	}

	builder := kaval.NewBuilder(options)
	result := builder.
		Enable("feature").
		Labeled("name", "john").
		LabeledList("tags", "dev", "prod").
		LabeledDict("settings", "theme", "dark", "fontSize", 14).
		String()

	fmt.Println(result)
}

// Output: ^feature, name = john, tags = dev; prod, settings = theme: dark; fontSize: 14
```


## 📜 License

This project is licensed under the 0BSD Licence — see the [LICENCE](LICENSE) file for details.

---

<p align="center">Made with ❤️ for structured data 🌟</p>
