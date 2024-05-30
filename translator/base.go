package translator

type Translator interface {
	Translate(text string, kwargs map[string]interface{}) (string, error)
	MapLanguageToCode(languages ...string) (string, string)
	SameSourceTarget() bool
	GetSupportedLanguages() interface{}
	IsLanguageSupported(language string) bool
	TranslateBatch(batch []string) ([]string, error)
	TranslateFile(path string) (string, error)
}
