# go-translate

Free to use Translation API for the most popular translation api's online.

## Installation

```bash
go get github.com/misenkashari/go-translate
```

## Usage

Here is a basic example of how to use go-translate:

```go
package main

import (
    "fmt"
    "github.com/misenkashari/go-translate/translator"
)

func main() {
    t := translator.NewGoogleTranslator("en", "it", nil)
    result, err := t.Translate("Hello, world!")
    if err != nil {
        panic(err)
    }
    fmt.Println(result)

    // output
    //Ciao Mondo!
}
```

## Supported Translators

1. GoogleTranslator
2. AzureTranslator (API Key required)
3. ApertiumTranslator
4. LingueeTranslator (API Key required)
5. LibreTranslator
6. MyMemoryTranslator
7. DeepLTranslator (Free and API key present)
