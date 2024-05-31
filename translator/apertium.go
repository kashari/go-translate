package translator

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/kashari/go-translate/constants"
)

type ApertiumTranslator struct {
	baseURL string
	source  string
	target  string
	proxies *url.URL
	client  *http.Client
}

func NewApertiumTranslator(source, target string, proxies *url.URL) *ApertiumTranslator {
	return &ApertiumTranslator{
		baseURL: constants.BASE_URLS["APERTIUM"],
		source:  source,
		target:  target,
		proxies: proxies,
		client:  &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxies)}},
	}
}

func (a *ApertiumTranslator) Translate(text string) (string, error) {

	url := a.baseURL + "?langpair=" + a.source + "|" + a.target + "&q=" + url.QueryEscape(text)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		log.Panic(err)
	}

	defer resp.Body.Close()

	var response struct {
		ResponseData struct {
			TranslatedText string `json:"translatedText"`
		} `json:"responseData"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		log.Panic(err)
	}

	return response.ResponseData.TranslatedText, nil
}

func (bt *ApertiumTranslator) TranslateBatch(batch []string) ([]string, error) {
	var translated []string
	for _, text := range batch {
		translatedText, err := bt.Translate(text)
		if err != nil {
			return nil, err
		}
		translated = append(translated, translatedText)
	}
	return translated, nil
}

func (bt *ApertiumTranslator) TranslateFile(path string) (string, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	text := string(file)
	translatedText, err := bt.Translate(text)
	if err != nil {
		return "", err
	}

	return translatedText, nil
}
