package main

import (
	"flag"
	"fmt"

	googletranslate "github.com/misenkashari/go-translate/translator/google-translate"
)

func main() {
	text := flag.String("text", "", "Text to translate")
	from := flag.String("from", "en", "Locale to translate to")
	to := flag.String("to", "es", "Locale to translate to")
	flag.Parse()

	translator := googletranslate.NewGoogleTranslator(*from, *to, nil)

	translated, err := translator.Translate(*text)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(translated)
}
