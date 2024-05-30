package googletranslate

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/misenkashari/go-translate/bread"
	"github.com/misenkashari/go-translate/constants"
)

// GoogleTranslator represents a translator using Google Translate under the hood.
type GoogleTranslator struct {
	baseURL            string
	source             string
	target             string
	proxies            *url.URL
	elementTag         string
	elementQuery       map[string]string
	payloadKey         string
	altElementQuery    map[string]string
	urlParams          url.Values
	supportedLanguages map[string]string
	client             *http.Client
}

// NewGoogleTranslator creates a new instance of GoogleTranslator.
func NewGoogleTranslator(source, target string, proxies *url.URL) *GoogleTranslator {
	return &GoogleTranslator{
		baseURL:            constants.BASE_URLS["GOOGLE_TRANSLATE"],
		source:             source,
		target:             target,
		proxies:            proxies,
		elementTag:         "div",
		elementQuery:       map[string]string{"class": "t0"},
		payloadKey:         "q",
		altElementQuery:    map[string]string{"class": "result-container"},
		urlParams:          url.Values{},
		supportedLanguages: constants.GOOGLE_LANGUAGES_TO_CODES,
		client:             &http.Client{},
	}
}

// Translate translates the given text from the source language to the target language.
func (gt *GoogleTranslator) Translate(text string) (string, error) {
	if len(strings.TrimSpace(text)) == 0 || len(text) > 5000 {
		return "", errors.New("invalid input text")
	}

	if gt.source == gt.target {
		return text, nil
	}

	gt.urlParams.Set("tl", gt.target)
	gt.urlParams.Set("sl", gt.source)
	gt.urlParams.Set(gt.payloadKey, text)

	resp, err := gt.client.Get(gt.baseURL + "?" + gt.urlParams.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := bread.GetWithClient(gt.baseURL+"?"+gt.urlParams.Encode(), gt.client)
	if err != nil {
		return "", err
	}

	doc := bread.HTMLParse(body)
	element := doc.Find(gt.elementTag, "class", gt.elementQuery["class"])
	if element.Error != nil {
		element = doc.Find(gt.elementTag, "class", gt.altElementQuery["class"])
		if element.Error != nil {
			return "", errors.New("translation not found")
		}
	}

	translatedText := element.FullText()
	if strings.TrimSpace(translatedText) == strings.TrimSpace(text) {
		return text, nil
	}

	return translatedText, nil
}

// Translates the text from the given file path.
func (gt *GoogleTranslator) TranslateFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return gt.Translate(string(content))
}

// TranslateBatch translates a batch of texts.
func (gt *GoogleTranslator) TranslateBatch(batch []string) ([]string, error) {
	var translations []string
	for _, text := range batch {
		translated, err := gt.Translate(text)
		if err != nil {
			return nil, err
		}
		translations = append(translations, translated)
	}
	return translations, nil
}

// MapLanguageToCode maps languages to their corresponding codes
func (bt *GoogleTranslator) MapLanguageToCode(languages ...string) (string, string) {
	var mappedLanguages []string
	for _, language := range languages {
		if language == "auto" || contains(bt.supportedLanguages, language) {
			mappedLanguages = append(mappedLanguages, language)
		} else if code, ok := bt.supportedLanguages[language]; ok {
			mappedLanguages = append(mappedLanguages, code)
		} else {
			panic(fmt.Sprintf("No support for the provided language: %s", language))
		}
	}
	return mappedLanguages[0], mappedLanguages[1]
}

// Checks if a map contains a value
func contains(m map[string]string, value string) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// Checks if the source and target languages are the same
func (bt *GoogleTranslator) SameSourceTarget() bool {
	return bt.source == bt.target
}

// Returns the supported languages
func (bt *GoogleTranslator) GetSupportedLanguages() interface{} {
	return bt.supportedLanguages
}

// Checks if a language is supported
func (bt *GoogleTranslator) IsLanguageSupported(language string) bool {
	return language == "auto" || contains(bt.supportedLanguages, language) || bt.supportedLanguages[language] != ""
}
