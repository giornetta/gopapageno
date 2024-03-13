Go PAPAGENO
========

Go PAPAGENO (PArallel PArser GENeratOr) is a parallel parser generator based on Floyd's Operator Precedence Grammars.

It generates parallel Go parsers starting from a lexer and a grammar specification.
These specification files resemble Flex and Bison ones, although with some differences.

The generated parsers are self-contained and can be used without further effort.

This work is based on [Papageno](https://github.com/PAPAGENO-devels/papageno), a C parallel parser generator.

### Installation
```
go get github.com/giornetta/gopapageno
```

### Parser generator example

```go
package main

import (
	"github.com/giornetta/gopapageno/generator"
)

func main() {
	generator.Generate("languages/arithmetic/lexer/arith.l", "languages/arithmetic/parser/arith.g", "languages/arithmetic")
}
```

### Parser usage example

```go
package main

import (
    "fmt"

    "github.com/giornetta/gopapageno/languages/arithmetic"
)

func main() {
    root, err := arithmetic.ParseFile("expression.txt", 2)
    
    if err == nil {
        fmt.Printf("Result: %d\n", *root.Value.(*int64))
    } else {
        fmt.Printf("Parse failed!")
        fmt.Println(err.Error())
    }
}
```

### Authors and Contributors

 * Simone Guidi <simone.guidi@mail.polimi.it>