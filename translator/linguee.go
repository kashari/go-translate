package translator

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/kashari/go-translate/bread"
	"github.com/kashari/go-translate/constants"
	errs "github.com/kashari/go-translate/errors"
)

type LingueeTranslator struct {
	baseURL            string
	source             string
	target             string
	elementTag         string
	elementQuery       map[string]string
	payloadKey         string
	proxies            *url.URL
	urlParams          url.Values
	supportedLanguages map[string]string
	client             *http.Client
}

func NewLingueeTranslator(source, target string, proxies *url.URL) *LingueeTranslator {
	return &LingueeTranslator{
		baseURL:    constants.BASE_URLS["LINGUEE"],
		source:     source,
		target:     target,
		elementTag: "a",
		elementQuery: map[string]string{
			"class": "dictLink featured",
		},
		proxies:            proxies,
		payloadKey:         "source",
		urlParams:          url.Values{},
		supportedLanguages: constants.LINGUEE_LANGUAGES_TO_CODES,
		client:             &http.Client{},
	}
}

func (bt *LingueeTranslator) SameSourceTarget() bool {
	return bt.source == bt.target
}

func (lt *LingueeTranslator) Translate(word string) (interface{}, error) {
	if lt.SameSourceTarget() || isEmpty(word) {
		return word, nil
	}

	if isInputValid(word, 50) {
		url := fmt.Sprintf("%s%s-%s/search/?source=%s&query=%s", lt.baseURL, lt.source, lt.target, lt.source, url.QueryEscape(word))

		response, err := bread.GetWithClient(url, lt.client)
		if err != nil {
			return nil, err
		}

		statusCode := 200
		if statusCode == 429 {
			errs.TooManyRequestsError()
			return nil, nil
		}

		root := bread.HTMLParse(response)
		if root.Error != nil {
			return nil, root.Error
		}

		log.Println(root)

		var elements []bread.Root
		for key, value := range lt.elementQuery {
			elements = root.FindAll(lt.elementTag, key, value)
		}

		if len(elements) == 0 {
			errs.TranslationNotFoundError()
			return nil, nil
		}

		var filteredElements []string
		for _, el := range elements {
			pronounElement := el.Find("span", "class", "placeholder")
			var pronoun string
			if pronounElement.Error == nil {
				pronoun = pronounElement.Text()
			}
			filteredElements = append(filteredElements, strings.ReplaceAll(el.Text(), pronoun, ""))
		}

		if len(filteredElements) == 0 {
			errs.TranslationNotFoundError()
			return nil, nil
		}

		return filteredElements, nil
	}
	return nil, fmt.Errorf("invalid input word")
}

func isInputValid(word string, maxChars int) bool {
	return len(word) > 0 && len(word) <= maxChars
}

func isEmpty(word string) bool {
	return len(word) == 0
}
