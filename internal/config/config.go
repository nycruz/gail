package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/nycruz/gail/internal/models"
)

// Config holds configuration data needed by the application.
type Config struct {
	Model         models.Model
	ModelMaxToken models.Token
	ModelAPIKey   string
	ConfigDir     string
}

const (
	envOpenAIAPIKey = "OPENAI_API_KEY"
	envClaudeAPIKey = "CLAUDE_API_KEY"
	configDirName   = ".config/gail"
	configFileExt   = "toml"
)

// New initializes a new Config struct based on the provided model flag and configures the necessary files.
func New(modelFlag, validationsFilename, assistantsFilename string) (*Config, error) {
	configDirPath, err := createConfigDir(configDirName)
	if err != nil {
		return nil, fmt.Errorf("failed to create the '%s' directory: %w", configDirName, err)
	}

	if err := createConfigFiles(configDirPath, validationsFilename, assistantsFilename); err != nil {
		return nil, err
	}

	modelName, maxTokens, apiKey, err := selectModelConfig(modelFlag)
	if err != nil {
		return nil, err
	}

	return &Config{
		Model:         modelName,
		ModelMaxToken: maxTokens,
		ModelAPIKey:   apiKey,
		ConfigDir:     configDirPath,
	}, nil
}

func selectModelConfig(modelFlag string) (models.Model, models.Token, string, error) {
	var modelName models.Model
	var maxTokens models.Token
	var apiKey string

	switch modelFlag {
	case models.ModelClaude:
		modelName = models.ModelClaudeName
		maxTokens = models.ModelClaudeMaxTokens
		apiKey = os.Getenv(envClaudeAPIKey)
		if apiKey == "" {
			return modelName, maxTokens, apiKey, fmt.Errorf("environment variable '%s' not set for model '%s'", envClaudeAPIKey, models.ModelClaude)
		}
	case models.ModelGPT:
		modelName = models.ModelGPTName
		maxTokens = models.ModelGPTMaxTokens
		apiKey = os.Getenv(envOpenAIAPIKey)
		if apiKey == "" {
			return modelName, maxTokens, apiKey, fmt.Errorf("environment variable '%s' not set for model '%s'", envOpenAIAPIKey, models.ModelGPT)
		}
	// case models.ModelGPTo1:
	// 	modelName = models.ModelGPTo1Name
	// 	maxTokens = models.ModelGPTo1MaxTokens
	// 	apiKey = os.Getenv(envOpenAIAPIKey)
	// 	if apiKey == "" {
	// 		return modelName, maxTokens, apiKey, fmt.Errorf("environment variable '%s' not set for model '%s'", envOpenAIAPIKey, models.ModelGPTo1)
	// 	}
	// case models.ModelGPTo1Mini:
	// 	modelName = models.ModelGPTo1MiniName
	// 	maxTokens = models.ModelGPTo1MiniMaxTokens
	// 	apiKey = os.Getenv(envOpenAIAPIKey)
	// 	if apiKey == "" {
	// 		return modelName, maxTokens, apiKey, fmt.Errorf("environment variable '%s' not set for model '%s'", envOpenAIAPIKey, models.ModelGPTo1Mini)
	// 	}
	default:
		return modelName, maxTokens, apiKey, fmt.Errorf("invalid model flag '%s'. Use one of ['%s', '%s']", modelFlag, models.ModelClaude, models.ModelGPT)
	}

	return modelName, maxTokens, apiKey, nil
}

func createConfigFiles(configDir, validationsFilename, assistantsFilename string) error {
	if err := createConfigFile(configDir, validationsFilename); err != nil {
		return err
	}

	if err := createConfigFile(configDir, assistantsFilename); err != nil {
		return err
	}

	return nil
}

func createConfigFile(configDir, filename string) error {
	sourceFilePath := filepath.Join(getAssetsDir(), fmt.Sprintf("%s.%s", filename, configFileExt))
	destFilePath := filepath.Join(configDir, fmt.Sprintf("%s.%s", filename, configFileExt))

	if _, err := os.Stat(destFilePath); err == nil {
		// File already exists, skip creation.
		return nil
	}

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil {
		return fmt.Errorf("failed to open '%s': %w", sourceFilePath, err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(destFilePath)
	if err != nil {
		return fmt.Errorf("failed to create the file in '%s': %w", destFilePath, err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("failed to copy file from '%s' to '%s': %w", sourceFilePath, destFilePath, err)
	}

	return nil
}

func createConfigDir(relativePath string) (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot get the user's home directory: %w", err)
	}

	configDirPath := filepath.Join(homeDir, relativePath)
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		if err := os.MkdirAll(configDirPath, 0755); err != nil {
			return "", fmt.Errorf("failed to create config directory '%s': %w", configDirPath, err)
		}
	}

	return configDirPath, nil
}

func getAssetsDir() string {
	currentDir, _ := os.Getwd()
	return filepath.Join(currentDir, "assets")
}
