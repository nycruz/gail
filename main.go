package main

import (
	"flag"
	"log"

	"log/slog"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/nycruz/gail/internal/assistant"
	"github.com/nycruz/gail/internal/config"
	"github.com/nycruz/gail/internal/logger"
	"github.com/nycruz/gail/internal/models"
	"github.com/nycruz/gail/internal/models/claude"
	"github.com/nycruz/gail/internal/models/gpt"
	"github.com/nycruz/gail/internal/tui"
	"github.com/nycruz/gail/internal/validator"
)

const (
	AppName             = "Gail"
	ValidationsFileName = "validations"
	AssistantsFileName  = "assistants"
)

func main() {
	modelFlag := flag.String("model", "gpt", "The model to use for the chat completion (e.g. gpt, claude)")
	logLevelFlag := flag.String("log-level", "info", "The log level to use for troubleshooting (e.g. debug, info, warn, error)")
	flag.Parse()

	var logLevel slog.Level
	switch *logLevelFlag {
	case "info":
		logLevel = slog.LevelInfo
	case "debug":
		logLevel = slog.LevelDebug
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	logger, err := logger.New(logLevel)
	if err != nil {
		log.Fatalf("ERROR: failed to instantiate 'logger': %v", err)
	}

	cfg, err := config.New(*modelFlag, ValidationsFileName, AssistantsFileName)
	if err != nil {
		log.Fatalf("ERROR: failed to instantiate 'config': %v", err)
	}

	validator, err := validator.New(logger, ValidationsFileName, cfg.ConfigDir)
	if err != nil {
		log.Fatalf("ERROR: failed to instantiate 'validator': %v", err)
	}

	assistant, err := assistant.New(logger, AssistantsFileName, cfg.ConfigDir)
	if err != nil {
		log.Fatalf("ERROR: failed to instantiate 'assistant': %v", err)
	}

	logger.Info(
		"Gail started. Loading LLM model...",
		slog.String("model", string(cfg.Model)),
		slog.Int("max_token", int(cfg.ModelMaxToken)),
	)

	var llm tui.LLM
	switch cfg.Model {
	case models.ModelGPTName:
		llm, err = gpt.New(logger, cfg.ModelAPIKey, cfg.Model, cfg.ModelMaxToken, AppName, validator)
		if err != nil {
			log.Fatalf("ERROR: failed to instantiate 'ChatGPT' model: %v", err)
		}
	case models.ModelClaudeName:
		llm, err = claude.New(logger, cfg.ModelAPIKey, cfg.Model, cfg.ModelMaxToken, AppName, validator)
		if err != nil {
			log.Fatalf("ERROR: failed to instantiate 'Claude' model: %v", err)
		}
	default:
		log.Fatalf("ERROR: failed to instantiate a model. '%s' is not supported", cfg.Model)
	}

	tui := tui.New(logger, llm, assistant)
	if err != nil {
		log.Fatalf("ERROR: failed to instantiate the Terminal User Interface: %v", err)
	}

	p := tea.NewProgram(tui, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("ERROR: failed to run the Terminal User Interface: %v", err)
	}
}
