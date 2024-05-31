package translator

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"

	"github.com/misenkashari/go-translate/constants"
)

type MyMemoryTranslator struct {
	baseURL string
	source  string
	target  string
	proxies *url.URL
	client  *http.Client
}

func NewMyMemoryTranslator(source, target string, proxies *url.URL) *MyMemoryTranslator {
	return &MyMemoryTranslator{
		baseURL: constants.BASE_URLS["MYMEMORY"],
		source:  source,
		target:  target,
		proxies: proxies,
		client:  &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxies)}},
	}
}

func (m *MyMemoryTranslator) Translate(text string) (string, error) {
	url := m.baseURL + "?langpair=" + m.source + "|" + m.target + "&q=" + url.QueryEscape(text)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Panic(err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
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
