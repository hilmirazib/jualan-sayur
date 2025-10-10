package validator

import (
	"errors"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	"github.com/rs/zerolog/log"
)

type Validator struct {
	Validator  *validator.Validate
	Translator ut.Translator
}

func NewValidator() *Validator {
	en := en.New()
	uni := ut.New(en, en)
	trans, found := uni.GetTranslator("en")
	if !found {
		log.Fatal().Msg("[NewValidator] Translator not found")
	}

	validate := validator.New()

	// Register default English translations
	if err := enTranslations.RegisterDefaultTranslations(validate, trans); err != nil {
		log.Fatal().Err(err).Msg("[NewValidator] Failed to register translations")
	}

	// Add custom validation messages
	validate.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is required", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	validate.RegisterTranslation("email", trans, func(ut ut.Translator) error {
		return ut.Add("email", "{0} must be a valid email address", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("email", fe.Field())
		return t
	})

	validate.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "{0} must be at least {1} characters", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("min", fe.Field(), fe.Param())
		return t
	})

	return &Validator{
		Validator:  validate,
		Translator: trans,
	}
}

func (v *Validator) Validate(i interface{}) error {
	err := v.Validator.Struct(i)

	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var errorMessages []string
			for _, e := range validationErrors {
				translatedMsg := e.Translate(v.Translator)
				log.Info().
					Str("field", e.Field()).
					Str("tag", e.Tag()).
					Str("value", e.Value().(string)).
					Str("message", translatedMsg).
					Msg("[Validate] Validation error")

				errorMessages = append(errorMessages, translatedMsg)
			}

			// Return the first error message
			if len(errorMessages) > 0 {
				return errors.New(errorMessages[0])
			}
		}

		// Fallback for non-validation errors
		return err
	}

	return nil
}

// ValidateAndGetErrors returns all validation errors as a slice
func (v *Validator) ValidateAndGetErrors(i interface{}) []string {
	err := v.Validator.Struct(i)

	if err != nil {
		var validationErrors validator.ValidationErrors
		if errors.As(err, &validationErrors) {
			var errorMessages []string
			for _, e := range validationErrors {
				translatedMsg := e.Translate(v.Translator)
				errorMessages = append(errorMessages, translatedMsg)
			}
			return errorMessages
		}
	}

	return nil
}

// ValidateField validates a single field
func (v *Validator) ValidateField(field interface{}, tag string) error {
	return v.Validator.Var(field, tag)
}
