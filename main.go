package main

import (
	"flag"
	"fmt"

	"github.com/kashari/go-translate/translator"
)

func main() {
	from := flag.String("from", "en", "source language")
	to := flag.String("to", "fr", "target language")
	text := flag.String("text", "Hello, World!", "text to translate")
	isFile := flag.String("file", "", "file to translate")
	flag.Parse()

	t := translator.NewGoogleTranslator(*from, *to, nil)

	if *isFile != "" {
		translated, err := t.TranslateFile(*isFile)
		if err != nil {
			panic(err)
		}

		fmt.Printf("Translated file named %s \n", translated.Name())
		return
	}

	translated, err := t.Translate(*text)
	if err != nil {
		panic(err)
	}

	fmt.Println(translated)
}
