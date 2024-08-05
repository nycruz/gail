package validator

import (
	"fmt"
	"regexp"

	"log/slog"

	"github.com/spf13/viper"
)

type Validator struct {
	Logger      *slog.Logger
	Validations []Validation
}

type ValidateConfig struct {
	Validations []Validation `mapstructure:"validation"`
}

type Validation struct {
	Name    string `mapstructure:"name"`
	Pattern string `mapstructure:"pattern"`
}

const (
	logPackageName = "validator"
)

// New creates a new Validator struct with the given validator path.
func New(logger *slog.Logger, validationsFilename string, configDirPath string) (*Validator, error) {
	fileExt := "toml"
	viper.SetConfigName(validationsFilename)
	viper.SetConfigType(fileExt)
	viper.AddConfigPath(configDirPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read the '%s.%s' config file: %w", validationsFilename, fileExt, err)
	}

	var vc *ValidateConfig
	err = viper.Unmarshal(&vc)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the '%s.%s' config file: %w", validationsFilename, fileExt, err)
	}

	v := &Validator{
		Logger:      logger,
		Validations: vc.Validations,
	}

	return v, nil
}

// Validate checks if the given user input matches any of the validation rules.
func (v *Validator) Validate(userInput string) (string, bool) {
	numValidations := len(v.Validations)
	v.Logger.Info(
		"Found validation rules.",
		slog.String("package", logPackageName),
		slog.Int("rules_count", numValidations),
	)

	for _, validation := range v.Validations {
		regex := regexp.MustCompile(validation.Pattern)
		if regex.MatchString(userInput) {
			v.Logger.Info(
				"Validation matched.",
				slog.String("package", logPackageName),
				slog.String("name", validation.Name),
			)

			displayErrorMessage := fmt.Sprintf("Your input contains Personal Identifiable Information: %s.\n Please try again!", validation.Name)
			return displayErrorMessage, false
		}
	}

	return "", true
}
