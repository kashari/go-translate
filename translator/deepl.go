package translator

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/kashari/go-translate/constants"
)

type DeepLTranslator struct {
	baseURL            string
	source             string
	target             string
	proxies            *url.URL
	urlParams          url.Values
	supportedLanguages map[string]string
	client             *http.Client
	apiKey             string
}

// Creates a new instance of DeepLTranslator.
// NOTEE: You need to provide an API key to use the API.
func NewDeepLTranslator(freeApi bool, apiKey string, source, target string, proxies *url.URL) *DeepLTranslator {
	if apiKey == "" {
		log.Panic("API key is required for the paid API")
	}
	var baseURL string
	var urlParams url.Values = url.Values{}

	if freeApi {
		baseURL = fmt.Sprintf("%s%s", constants.BASE_URLS["DEEPL_FREE"], "translate")
	} else {
		baseURL = fmt.Sprintf("%s%s", constants.BASE_URLS["DEEPL"], "translate")
	}

	urlParams.Add("source_lang", source)
	urlParams.Add("target_lang", target)

	return &DeepLTranslator{
		baseURL:            baseURL,
		source:             source,
		target:             target,
		proxies:            proxies,
		urlParams:          urlParams,
		supportedLanguages: constants.DEEPL_LANGUAGE_TO_CODE,
		client:             &http.Client{},
		apiKey:             apiKey,
	}
}

func (d *DeepLTranslator) Translate(text string) (string, error) {
	d.urlParams.Set("text", text)
	// send a request with all the params
	req, err := http.NewRequest("POST", d.baseURL, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Authorization", fmt.Sprintf("DeepL-Auth-Key %s", d.apiKey))

	req.URL.RawQuery = d.urlParams.Encode()

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	// read the response
	translatedText, err := readResponse(resp)
	if err != nil {
		return "", err
	}

	return translatedText, nil
}

func readResponse(resp *http.Response) (string, error) {
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to translate text: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}

func (d *DeepLTranslator) TranslateBatch(texts []string) ([]string, error) {
	var wg sync.WaitGroup
	translations := make([]string, len(texts))
	ch := make(chan struct {
		index int
		text  string
		err   error
	}, len(texts))

	for i, text := range texts {
		wg.Add(1)
		go func(i int, text string) {
			defer wg.Done()
			translated, err := d.Translate(text)
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

func (d *DeepLTranslator) TranslateFile(path string) (string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return d.Translate(string(text))
}

func (d *DeepLTranslator) MapLanguageToCode(languages ...string) (string, string) {
	var mappedLanguages []string
	for _, language := range languages {
		if language == "auto" || contains(d.supportedLanguages, language) {
			mappedLanguages = append(mappedLanguages, language)
		} else if code, ok := d.supportedLanguages[language]; ok {
			mappedLanguages = append(mappedLanguages, code)
		} else {
			panic(fmt.Sprintf("No support for the provided language: %s", language))
		}
	}
	return mappedLanguages[0], mappedLanguages[1]
}

func (d *DeepLTranslator) SameSourceTarget() bool {
	return d.source == d.target
}

func (d *DeepLTranslator) GetSupportedLanguages() interface{} {
	return d.supportedLanguages
}

func (d *DeepLTranslator) IsLanguageSupported(language string) bool {
	return language == "auto" || contains(d.supportedLanguages, language) || d.supportedLanguages[language] != ""
}
