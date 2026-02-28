package transport

import (
	"strings"

	"github.com/azizndao/grouter"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/es"
	"github.com/go-playground/locales/fr"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
	es_translations "github.com/go-playground/validator/v10/translations/es"
	fr_translations "github.com/go-playground/validator/v10/translations/fr"
)

// initializeTranslators sets up the universal translator with multiple locales
func initializeTranslators(validate *validator.Validate) map[string]ut.Translator {
	// Create locales
	enLocale := en.New()
	esLocale := es.New()
	frLocale := fr.New()

	// Create universal translator
	uni := ut.New(enLocale, enLocale, esLocale, frLocale)

	translators := make(map[string]ut.Translator)

	// Setup English translations
	setupEnglishTranslations(uni, validate, translators)

	// Setup Spanish translations
	setupSpanishTranslations(uni, validate, translators)

	// Setup French translations
	setupFrenchTranslations(uni, validate, translators)

	return translators
}

// setupEnglishTranslations configures English locale translations
func setupEnglishTranslations(uni *ut.UniversalTranslator, validate *validator.Validate, translators map[string]ut.Translator) {
	enTrans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(validate, enTrans)

	// Map multiple English variants to the same translator
	englishVariants := []string{"en", "en-US", "en-GB", "en-CA", "en-AU"}
	for _, variant := range englishVariants {
		translators[variant] = enTrans
	}
}

// setupSpanishTranslations configures Spanish locale translations
func setupSpanishTranslations(uni *ut.UniversalTranslator, validate *validator.Validate, translators map[string]ut.Translator) {
	esTrans, _ := uni.GetTranslator("es")
	es_translations.RegisterDefaultTranslations(validate, esTrans)

	// Map multiple Spanish variants to the same translator
	spanishVariants := []string{"es", "es-ES", "es-MX", "es-AR", "es-CO"}
	for _, variant := range spanishVariants {
		translators[variant] = esTrans
	}
}

// setupFrenchTranslations configures French locale translations
func setupFrenchTranslations(uni *ut.UniversalTranslator, validate *validator.Validate, translators map[string]ut.Translator) {
	frTrans, _ := uni.GetTranslator("fr")
	fr_translations.RegisterDefaultTranslations(validate, frTrans)

	// Map multiple French variants to the same translator
	frenchVariants := []string{"fr", "fr-FR", "fr-CA", "fr-BE", "fr-CH"}
	for _, variant := range frenchVariants {
		translators[variant] = frTrans
	}
}

// detectLocale detects the user's preferred locale from the Accept-Language header
func (t *Default) detectLocale(c *grouter.Ctx) string {
	acceptLang := c.Get(HeaderAcceptLanguage)
	if acceptLang == "" {
		return DefaultLocale
	}

	return t.parseAcceptLanguage(acceptLang)
}

// parseAcceptLanguage parses Accept-Language header and returns best matching locale
func (t *Default) parseAcceptLanguage(acceptLang string) string {
	// Parse Accept-Language header format: "en-US,en;q=0.9,es;q=0.8,fr;q=0.7"
	languages := strings.SplitSeq(acceptLang, ",")

	for lang := range languages {
		// Remove quality factors and whitespace
		lang = strings.TrimSpace(strings.Split(lang, ";")[0])

		// Check exact match first
		if _, exists := t.translators[lang]; exists {
			return lang
		}

		// Try base language fallback (e.g., "en" from "en-US")
		if baseLang := strings.Split(lang, "-")[0]; baseLang != lang {
			if _, exists := t.translators[baseLang]; exists {
				return baseLang
			}
		}
	}

	return DefaultLocale
}

// getTranslator returns the translator for the specified locale
func (t *Default) getTranslator(locale string) ut.Translator {
	// Try exact match first
	if translator, exists := t.translators[locale]; exists {
		return translator
	}

	// Try base language
	if baseLang := strings.Split(locale, "-")[0]; baseLang != locale {
		if translator, exists := t.translators[baseLang]; exists {
			return translator
		}
	}

	// Default to English
	return t.translators[DefaultLocale]
}
