package config

import (
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

type TargetRef struct {
	ProjectID     string
	EnvironmentID string
}

func (r TargetRef) ID() string {
	return TargetID(r.ProjectID, r.EnvironmentID)
}

func TargetID(projectID, environmentID string) string {
	return projectID + ":" + environmentID
}

func ParseTargetID(value string) (TargetRef, bool) {
	if strings.Count(value, ":") != 1 {
		return TargetRef{}, false
	}
	projectID, environmentID, _ := strings.Cut(value, ":")
	projectID = strings.TrimSpace(projectID)
	environmentID = strings.TrimSpace(environmentID)
	if !validID(projectID) || !validID(environmentID) {
		return TargetRef{}, false
	}
	return TargetRef{ProjectID: projectID, EnvironmentID: environmentID}, true
}

func (c *Config) TargetIDs() []string {
	targets := make([]string, 0, c.EnvironmentCount())
	for projectID, project := range c.Projects {
		for environmentID := range project.Environments {
			targets = append(targets, TargetID(projectID, environmentID))
		}
	}
	sort.Strings(targets)
	return targets
}

func (c *Config) Target(ref TargetRef) (Project, Environment, bool) {
	project, projectExists := c.Projects[ref.ProjectID]
	if !projectExists {
		return Project{}, Environment{}, false
	}
	environment, environmentExists := project.Environments[ref.EnvironmentID]
	return project, environment, environmentExists
}

func (c *Config) FindTargetByDirectory(directory string, preferredEnvironments ...string) (TargetRef, bool) {
	matches := c.targetsByDirectory(directory)
	if len(matches) == 1 {
		return matches[0], true
	}
	for _, preferred := range preferredEnvironments {
		var selected TargetRef
		count := 0
		for _, match := range matches {
			if match.EnvironmentID == preferred {
				selected = match
				count++
			}
		}
		if count == 1 {
			return selected, true
		}
	}
	return TargetRef{}, false
}

func (c *Config) HasTargetDirectory(directory string) bool {
	return len(c.targetsByDirectory(directory)) > 0
}

func (c *Config) targetsByDirectory(directory string) []TargetRef {
	directory = filepath.Clean(directory)
	matches := make([]TargetRef, 0)
	for projectID, project := range c.Projects {
		for environmentID, environment := range project.Environments {
			if filepath.Clean(environment.Directory) == directory {
				matches = append(matches, TargetRef{ProjectID: projectID, EnvironmentID: environmentID})
			}
		}
	}
	return matches
}

func IsValidSSHTarget(value string) bool {
	value = strings.TrimSpace(value)
	if value == "" || strings.HasPrefix(value, "-") || strings.Contains(value, "/") || strings.ContainsFunc(value, func(character rune) bool {
		return unicode.IsSpace(character) || unicode.IsControl(character)
	}) {
		return false
	}
	if strings.Count(value, "@") > 1 || strings.HasPrefix(value, "@") || strings.HasSuffix(value, "@") {
		return false
	}
	host := value
	if separator := strings.LastIndex(value, "@"); separator >= 0 {
		host = value[separator+1:]
	}
	return !strings.Contains(host, ":") || strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]")
}

func (t Target) String() string {
	return t.Host + ":" + t.Directory
}
