package translator

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"

	"github.com/kashari/go-translate/bread"
	"github.com/kashari/go-translate/constants"
	errs "github.com/kashari/go-translate/errors"
)

// Represents a translator using Google Translate under the hood.
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

// Creates a new instance of GoogleTranslator.
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

// Translates the given text from the source language to the target language.
func (gt *GoogleTranslator) Translate(text string) (string, error) {
	if len(strings.TrimSpace(text)) == 0 || len(text) > 5000 {
		errs.TooLongTextError()
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
func (gt *GoogleTranslator) TranslateFile(path string) (*os.File, error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		log.Panic("File with given path not found")
	}

	// Since Google supports up to 5000 chars we split the content into multiple 5000 char chunks
	var chunks []string
	for i := 0; i < len(bytes); i += 5000 {
		end := i + 5000
		if end > len(bytes) {
			end = len(bytes)
		}
		chunks = append(chunks, string(bytes[i:end]))
	}

	translatedChunks := make([]string, len(chunks))
	var wg sync.WaitGroup
	var mu sync.Mutex
	errChan := make(chan error, 1) // Use a buffered channel with capacity 1 for errors

	for i, chunk := range chunks {
		wg.Add(1)
		go func(i int, chunk string) {
			defer wg.Done()

			// Create a copy of the urlParams for each goroutine
			urlParamsCopy := make(url.Values)
			for k, v := range gt.urlParams {
				urlParamsCopy[k] = v
			}

			translated, err := gt.TranslateWithParams(chunk, urlParamsCopy)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				return
			}
			mu.Lock()
			translatedChunks[i] = translated
			mu.Unlock()
		}(i, chunk)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Check if there was any error
	if err := <-errChan; err != nil {
		return nil, err
	}

	translated := strings.Join(translatedChunks, "")
	filename := fmt.Sprintf("translated_%s", GetFileNameFromPath(path))
	file, err := os.Create(filename)
	if err != nil {
		log.Println("Error: ", err)
		return nil, err
	}
	defer file.Close()

	_, err = file.WriteString(translated)
	if err != nil {
		return nil, err
	}

	return file, nil
}

// Translates the given text with the provided URL parameters.
func (gt *GoogleTranslator) TranslateWithParams(text string, urlParams url.Values) (string, error) {
	if len(strings.TrimSpace(text)) == 0 || len(text) > 5000 {
		errs.TooLongTextError()
		return "", errors.New("invalid input text")
	}

	if gt.source == gt.target {
		return text, nil
	}

	urlParams.Set("tl", gt.target)
	urlParams.Set("sl", gt.source)
	urlParams.Set(gt.payloadKey, text)

	resp, err := gt.client.Get(gt.baseURL + "?" + urlParams.Encode())
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := bread.GetWithClient(gt.baseURL+"?"+urlParams.Encode(), gt.client)
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

// Translates a batch of texts.
func (gt *GoogleTranslator) TranslateBatch(batch []string) ([]string, error) {
	var wg sync.WaitGroup
	translations := make([]string, len(batch))
	ch := make(chan struct {
		index int
		text  string
		err   error
	}, len(batch))

	for i, text := range batch {
		if len(text) > 5000 {
			// split the text in chunks of 5000 characters
			var chunks []string
			for i := 0; i < len(text); i += 5000 {
				end := i + 5000
				if end > len(text) {
					end = len(text)
				}
				chunks = append(chunks, text[i:end])
			}
			wg.Add(len(chunks))
			for j, chunk := range chunks {
				go func(i, j int, text string) {
					defer wg.Done()
					translated, err := gt.Translate(text)
					ch <- struct {
						index int
						text  string
						err   error
					}{i, translated, err}
				}(i, j, chunk)
			}
			continue
		}

		wg.Add(1)
		go func(i int, text string) {
			defer wg.Done()
			translated, err := gt.Translate(text)
			ch <- struct {
				index int
				text  string
				err   error
			}{i, translated, err}
		}(i, text)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for result := range ch {
		if result.err != nil {
			return nil, result.err
		}
		translations[result.index] = result.text
	}

	return translations, nil
}

// Maps languages to their corresponding codes
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

func GetFileNameFromPath(path string) string {
	name := strings.TrimRight(path, "/")
	name = strings.Split(name, "/")[len(strings.Split(name, "/"))-1]
	return name
}
