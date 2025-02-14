package assistant

import (
	"fmt"
	"log/slog"

	"github.com/spf13/viper"
)

type Assistant struct {
	Logger *slog.Logger
	Roles  []Role
	Skills []Skill
}

type AssistantConfig struct {
	Roles  []Role  `mapstructure:"role"`
	Skills []Skill `mapstructure:"skill"`
}

type Role struct {
	ID      string `mapstructure:"id"`
	Name    string `mapstructure:"name"`
	Persona string `mapstructure:"persona"`
}

type Skill struct {
	ID          string   `mapstructure:"id"`
	Instruction string   `mapstructure:"instruction"`
	Description string   `mapstructure:"description"`
	RoleIDs     []string `mapstructure:"roleIDs"`
}

// New creates a new Assistant instance.
func New(logger *slog.Logger, assistantsFilename string, configDirPath string) (*Assistant, error) {
	fileExt := "toml"
	viper.SetConfigName(assistantsFilename)
	viper.SetConfigType(fileExt)
	viper.AddConfigPath(configDirPath)
	err := viper.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read the '%s.%s' config file: %w", assistantsFilename, fileExt, err)
	}

	var ac *AssistantConfig
	err = viper.Unmarshal(&ac)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal the '%s.%s' config: %w", assistantsFilename, fileExt, err)
	}

	a := &Assistant{
		Logger: logger,
		Roles:  ac.Roles,
		Skills: ac.Skills,
	}

	return a, nil
}

// DefaultRole returns the default role.
func (a *Assistant) DefaultRole() Role {
	for _, role := range a.Roles {
		if role.ID == "swe" {
			return role
		}
	}
	return Role{}
}

// DefaultSkill returns the default skill.
func (a *Assistant) DefaultSkill() Skill {
	for _, skill := range a.Skills {
		if skill.ID == "default" {
			return skill
		}
	}
	return Skill{}
}

// FindRoleByID returns a role by its ID.
func (a *Assistant) FindRoleByID(id string) Role {
	for _, role := range a.Roles {
		if role.ID == id {
			return role
		}
	}
	return Role{}
}

// GetRoleSkills returns a list of skills for a given role ID.
func (a *Assistant) GetRoleSkills(roleID string) []Skill {
	var skills []Skill
	for _, skill := range a.Skills {
		for _, id := range skill.RoleIDs {
			if id == roleID {
				skills = append(skills, skill)
			}
		}
	}
	return skills
}
