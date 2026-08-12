package presentation

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
)

func RenderProjects(cfg *config.Config, color bool, translator i18n.Translator) string {
	if cfg.EnvironmentCount() == 0 {
		return translator.T(i18n.NoTargets)
	}
	accent := lipgloss.NewStyle()
	if color {
		accent = accent.Foreground(lipgloss.Color("42")).Bold(true)
	}
	var output strings.Builder
	index := 1
	projectIDs := keys(cfg.Projects)
	for _, projectID := range projectIDs {
		project := cfg.Projects[projectID]
		for _, environmentID := range keys(project.Environments) {
			environment := project.Environments[environmentID]
			fmt.Fprintf(&output, "%s %s · %s\n", accent.Render(fmt.Sprintf("[%d]", index)), project.Name, environment.Name)
			fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldIdentifier), config.TargetID(projectID, environmentID)))
			fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldLocal), environment.Directory))
			fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldRemote), environment.Target.String()))
			if len(environment.Git.AllowedBranches) == 0 {
				fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldBranches), translator.T(i18n.AnyBranchWarning)))
			} else {
				fmt.Fprintf(&output, "    %s\n", translator.T(i18n.FieldLine, translator.T(i18n.FieldBranches), strings.Join(environment.Git.AllowedBranches, ", ")))
			}
			output.WriteByte('\n')
			index++
		}
	}
	return output.String()
}

func keys[T any](values map[string]T) []string {
	result := make([]string, 0, len(values))
	for key := range values {
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}
