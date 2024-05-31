package translator

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
)

type AzureTranslator struct {
	baseURL string
	source  string
	target  string
	proxies *url.URL
	client  *http.Client
	apiKey  string
	region  string
}

func NewAzureTranslator(source, target string, proxies *url.URL, apiKey, region string) *AzureTranslator {
	return &AzureTranslator{
		baseURL: "https://api.cognitive.microsofttranslator.com/translate?api-version=3.0",
		source:  source,
		target:  target,
		proxies: proxies,
		client:  &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxies)}},
		apiKey:  apiKey,
		region:  region,
	}
}

func (a *AzureTranslator) Translate(text string) (string, error) {
	u, _ := url.Parse(a.baseURL)
	q := u.Query()
	q.Add("from", a.source)
	q.Add("to", a.target)
	u.RawQuery = q.Encode()

	body := []struct {
		Text string
	}{
		{Text: text},
	}
	b, _ := json.Marshal(body)

	req, err := http.NewRequest("POST", u.String(), bytes.NewBuffer(b))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Add("Ocp-Apim-Subscription-Key", a.apiKey)
	req.Header.Add("Ocp-Apim-Subscription-Region", a.region)
	req.Header.Add("Content-Type", "application/json")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	var result interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Fatal(err)
	}

	translations := result.([]interface{})
	translatedText := translations[0].(map[string]interface{})["translations"].([]interface{})[0].(map[string]interface{})["text"].(string)
	return translatedText, nil
}

func (a *AzureTranslator) TranslateBatch(texts []string) ([]string, error) {
	var translatedTexts []string
	for _, text := range texts {
		translatedText, err := a.Translate(text)
		if err != nil {
			return nil, err
		}
		translatedTexts = append(translatedTexts, translatedText)
	}
	return translatedTexts, nil
}

func (a *AzureTranslator) TranslateFile(path string) (string, error) {
	text, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	return a.Translate(string(text))
}
