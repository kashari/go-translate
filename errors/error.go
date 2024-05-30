package errors

import (
	"errors"
	"log"
)

func RequestError() {
	err := errors.New("request error, please try again later")
	log.Panic(err)
}

func TooManyRequestsError() {
	err := errors.New("too many requests, please try again later")
	log.Panic(err)
}

func TranslationNotFoundError() {
	err := errors.New("translation not found")
	log.Panic(err)
}

func InvalidSourceOrTargetLanguageError() {
	err := errors.New("invalid source or target language")
	log.Panic(err)
}

func LanguageNotSupportedExceptionError() {
	err := errors.New("language not supported")
	log.Panic(err)
}

func InvalidPayloadKeyError() {
	err := errors.New("invalid payload key")
	log.Panic(err)
}
