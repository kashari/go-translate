package deepl

import (
	"net/http"
	"net/url"

	"github.com/misenkashari/go-translate/constants"
)

type DeepLTranslator struct {
	baseURL         string
	source          string
	target          string
	proxies         *url.URL
	apiKey          string
	elementTag      string
	elementQuery    map[string]string
	payloadKey      string
	altElementQuery map[string]string
	urlParams       url.Values
	useFreeApi      bool
	client          *http.Client
}

func NewDeepLTranslator(source, target, apiKey string, proxies *url.URL) *DeepLTranslator {
	return &DeepLTranslator{
		baseURL:         "https://api.deepl.com/v2/translate",
		source:          source,
		target:          target,
		proxies:         proxies,
		apiKey:          constants.DEEPL_ENV_VAR,
		elementTag:      "div",
		elementQuery:    map[string]string{"class": "t0"},
		payloadKey:      "text",
		altElementQuery: map[string]string{"class": "result-container"},
		urlParams:       url.Values{},
		useFreeApi:      true,
		client:          &http.Client{},
	}
}
