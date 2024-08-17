package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nycruz/gail/internal/models"
)

type Config struct {
	Model         models.Model
	ModelMaxToken models.Token
	ModelAPIKey   string
	ConfigDir     string
}

const (
	ENV_OPENAI_API_KEY = "OPENAI_API_KEY"
	ENV_CLAUDE_API_KEY = "CLAUDE_API_KEY"
)

// New creates a new Config struct with the given model flag and configures the necessary files for the application.
func New(modelFlag string, validationsFilename string, assistantsFilename string) (*Config, error) {
	partialConfigDirPath := ".config/gail"
	configDirPath, err := createConfigDirectory(partialConfigDirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create the '%s' directory: %w", partialConfigDirPath, err)
	}

	if err := createConfigFiles(configDirPath, validationsFilename, assistantsFilename); err != nil {
		return nil, err
	}

	var modelName models.Model
	var maxTokens models.Token
	var apiKey string

	switch modelFlag {
	case models.ModelClaude:
		modelName = models.ModelClaudeName
		maxTokens = models.ModelClaudeMaxTokens
		apiKey = os.Getenv(ENV_CLAUDE_API_KEY)
		if apiKey == "" {
			return nil, fmt.Errorf("'%s' environment variable not found in order to use model '%s'", ENV_CLAUDE_API_KEY, models.ModelClaude)
		}
	case models.ModelGPT:
		modelName = models.ModelGPTName
		maxTokens = models.ModelGPT0125MaxTokens
		apiKey = os.Getenv(ENV_OPENAI_API_KEY)
		if apiKey == "" {
			return nil, fmt.Errorf("'%s' environment variable not found in order to use model '%s'", ENV_OPENAI_API_KEY, models.ModelGPT)
		}
	default:
		return nil, fmt.Errorf("invalid model flag '%s'. Use one of ['%s','%s']", modelFlag, models.ModelClaude, models.ModelGPT)
	}

	return &Config{
		Model:         modelName,
		ModelMaxToken: maxTokens,
		ModelAPIKey:   apiKey,
		ConfigDir:     configDirPath,
	}, nil
}

// createConfigFiles creates the necessary config files for the application.
func createConfigFiles(configDirPath string, validationsFilename string, assistantsFilename string) error {
	configFileExtention := "toml"
	validationsFullFilename := fmt.Sprintf("%s.%s", validationsFilename, configFileExtention)
	assistantsFullFilename := fmt.Sprintf("%s.%s", assistantsFilename, configFileExtention)

	if err := createValidatorConfigFile(configDirPath, validationsFullFilename); err != nil {
		return fmt.Errorf("failed to create the '%s' file: %w", validationsFullFilename, err)
	}

	if err := createAssistantsConfigFile(configDirPath, assistantsFullFilename); err != nil {
		return fmt.Errorf("failed to create the '%s' config file: %w", assistantsFullFilename, err)
	}

	return nil
}

// createConfigDirectory creates the '.config/' directory in the user's home directory.
func createConfigDirectory(partialconfigDirPath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get the user's home directory: %w", err)
	}

	configDirPath := fmt.Sprintf("%s/%s", homeDir, partialconfigDirPath)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		err = os.Mkdir(configDirPath, 0755)
		if err != nil {
			return "", err
		}
	}

	return configDirPath, nil
}

// createValidatorConfigFile creates the 'validations.toml' file in the config directory.
func createValidatorConfigFile(configDir string, validationsFullFilename string) error {
	currentDir, _ := os.Getwd()
	sourceFilePath := fmt.Sprintf("%s/%s/%s", currentDir, "assets", validationsFullFilename)
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", sourceFilePath, err)
	}
	defer sourceFile.Close()

	destinationFilePath := filepath.Join(configDir, filepath.Base(validationsFullFilename))
	// Check if the file already exists, if it does, return nil and do nothing
	if _, err := os.Stat(destinationFilePath); err == nil {
		return nil
	}

	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		return fmt.Errorf("failed to create the file in directory '%s': %w", destinationFilePath, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy the file from '%s' to '%s': %w", sourceFilePath, destinationFilePath, err)
	}
	return nil
}

// createAssistantsConfigFile creates the 'assistants.toml' file in the config directory.
func createAssistantsConfigFile(configDir string, assistantsFullFilename string) error {
	currentDir, _ := os.Getwd()
	sourceFilePath := fmt.Sprintf("%s/%s/%s", currentDir, "assets", assistantsFullFilename)
	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", sourceFilePath, err)
	}
	defer sourceFile.Close()

	destinationFilePath := filepath.Join(configDir, filepath.Base(assistantsFullFilename))
	// Check if the file already exists, if it does, return nil and do nothing
	if _, err := os.Stat(destinationFilePath); err == nil {
		return nil
	}

	destinationFile, err := os.Create(destinationFilePath)
	if err != nil {
		return fmt.Errorf("failed to create the file in directory '%s': %w", destinationFilePath, err)
	}
	defer destinationFile.Close()

	_, err = io.Copy(destinationFile, sourceFile)
	if err != nil {
		return fmt.Errorf("failed to copy the file from '%s' to '%s': %w", sourceFilePath, destinationFilePath, err)
	}
	return nil
}
