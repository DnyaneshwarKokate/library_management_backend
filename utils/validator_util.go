package utils

import (
	"encoding/base64"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	english "github.com/go-playground/validator/v10/translations/en"
	"github.com/sirupsen/logrus"
)

// InitializeValidator initializes the validator with custom translations and validation rules.
func InitializeValidator() (*validator.Validate, ut.Translator) {
	translator := en.New()
	uni := ut.New(translator, translator)

	trans, found := uni.GetTranslator("en")
	if !found {
		logrus.Error("Translator not found")
	}

	tokenValidator := validator.New()

	// Register default translations for English
	if err := english.RegisterDefaultTranslations(tokenValidator, trans); err != nil {
		logrus.Error(err)
	}

	// Register custom translations and validation rules here
	_ = tokenValidator.RegisterTranslation("required", trans, func(ut ut.Translator) error {
		return ut.Add("required", "{0} is required.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required", fe.Field())
		return t
	})

	_ = tokenValidator.RegisterTranslation("required_if", trans, func(ut ut.Translator) error {
		return ut.Add("required_if", "{0} is a required.", true) // see universal-translator for details
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T("required_if", fe.Field())
		return t
	})

	tokenValidator.RegisterTranslation("min", trans, func(ut ut.Translator) error {
		return ut.Add("min", "The {0} field must be at least {1} characters long.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
		return t
	})

	tokenValidator.RegisterTranslation("max", trans, func(ut ut.Translator) error {
		return ut.Add("max", "The {0} field cannot be longer than {1} characters.", true)
	}, func(ut ut.Translator, fe validator.FieldError) string {
		t, _ := ut.T(fe.Tag(), fe.Field(), fe.Param())
		return t
	})

	return tokenValidator, trans
}

// ValidateRequest performs validation on a request using the provided validator and translator.
func ValidateRequest(c *gin.Context, requestBody interface{}) interface{} {
	tokenValidator, trans := InitializeValidator() // Initialize the validator and translator

	tokenValidatorErr := tokenValidator.Struct(requestBody)

	if tokenValidatorErr != nil {
		var errorMessage string
		var counter int

		tokenValidatorErrors, ok := tokenValidatorErr.(validator.ValidationErrors)
		if !ok {
			logrus.Error("Failed to assert ValidationErrors")
			return "Validation error"
		}

		size := len(tokenValidatorErrors)

		for _, e := range tokenValidatorErrors {
			counter++
			errorMessage = errorMessage + e.Translate(trans) // Use the provided translator
			if counter != size {
				errorMessage = errorMessage + "|"
			}
		}

		logrus.Error("errorMessage", errorMessage)
		return errorMessage
	}

	return nil
}

func IsValidName(name string) bool {
	for _, char := range name {
		if (char < 'A' || char > 'Z') && (char < 'a' || char > 'z') && char != ' ' {
			return false
		}
	}
	return true
}

func IsValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}
func IsValidMobile(mobile string) bool {
	mobileRegex := regexp.MustCompile(`^[0-9]{10}$`)
	return mobileRegex.MatchString(mobile)
}

func IsValidBase64(str string) bool {
	_, err := base64.StdEncoding.DecodeString(str)
	return err == nil
}

func NormalizeMobileNumber(mobile string) string {

	mobile = strings.TrimPrefix(mobile, "+91")
	mobile = strings.TrimSpace(mobile)

	// Check if it's now a valid 10-digit number
	if len(mobile) == 10 && regexp.MustCompile(`^\d{10}$`).MatchString(mobile) {
		return mobile
	}
	return mobile
}
