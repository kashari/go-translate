package translator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"sync"

	"github.com/kashari/go-translate/constants"
)

type LibreTranslator struct {
	baseURL string
	source  string
	target  string
	proxies *url.URL
	client  *http.Client
}

func NewLibreTranslator(source, target string, proxies *url.URL) *LibreTranslator {
	return &LibreTranslator{
		baseURL: constants.BASE_URLS["LIBRE_FREE"],
		source:  source,
		target:  target,
		proxies: proxies,
		client:  &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxies)}},
	}
}

func (l *LibreTranslator) Translate(text string) (string, error) {
	body := json.RawMessage(`{"q":"` + text + `","source":"` + l.source + `","target":"` + l.target + `"}`)

	req, err := http.NewRequest("POST", l.baseURL, bytes.NewBuffer(body))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := l.client.Do(req)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	var response struct {
		TranslatedText string `json:"translatedText"`
	}

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return "", err
	}

	return response.TranslatedText, nil
}

func (l *LibreTranslator) TranslateBatch(texts []string) ([]string, error) {
	var wg sync.WaitGroup
	translations := make([]string, len(texts))
	ch := make(chan struct {
		index int
		text  string
		err   error
	})

	for i, text := range texts {
		wg.Add(1)
		go func(i int, text string) {
			defer wg.Done()
			translated, err := l.Translate(text)
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

	for t := range ch {
		if t.err != nil {
			return nil, t.err
		}
		translations[t.index] = t.text
	}

	return translations, nil
}

func (l *LibreTranslator) TranslateFile(path string) (string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return l.Translate(string(text))
}
