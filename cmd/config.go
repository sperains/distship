package cmd

import (
	"bufio"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
)

func (a *app) newConfigCommand() *cobra.Command {
	command := &cobra.Command{
		Use:   "config",
		Short: a.translator.T(i18n.ConfigShort),
	}
	command.AddCommand(&cobra.Command{
		Use:   "validate",
		Short: a.translator.T(i18n.ConfigValidateShort),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, path, err := a.loadConfig()
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
			}
			_, err = fmt.Fprintf(a.out, "✓ %s\n\n  %s\n  %s\n  %s\n", a.translator.T(i18n.ConfigValid), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldPath), path), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldProjects), len(cfg.Projects)), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldTargets), cfg.EnvironmentCount()))
			return err
		},
	}, a.newConfigRemoveCommand())
	return command
}

func (a *app) newConfigRemoveCommand() *cobra.Command {
	var yes bool
	command := &cobra.Command{
		Use:     "remove <target-id>",
		Short:   a.translator.T(i18n.ConfigRemoveShort),
		Example: "  distship config remove ipd:test",
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return i18n.NewError(i18n.InvalidTargetID)
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			target, ok := config.ParseTargetID(args[0])
			if !ok {
				return i18n.NewError(i18n.InvalidTargetID)
			}
			return a.runConfigRemove(target, yes)
		},
	}
	command.Flags().BoolVar(&yes, "yes", false, a.translator.T(i18n.FlagYes))
	return command
}

func (a *app) runConfigRemove(target config.TargetRef, yes bool) error {
	cfg, path, err := a.loadConfig()
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	project, environment, exists := cfg.Target(target)
	if !exists {
		return i18n.NewError(i18n.TargetNotFound, target.ID(), strings.Join(cfg.TargetIDs(), "\n  "))
	}

	lines := []string{
		a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldProject), project.Name),
		a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldEnvironment), environment.Name),
		a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldIdentifier), target.ID()),
		a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldLocalDirectory), environment.Directory),
		a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldDeployTarget), environment.Target.String()),
	}
	fmt.Fprintf(a.out, "%s\n\n  %s\n\n%s\n", a.translator.T(i18n.RemovePreviewTitle), strings.Join(lines, "\n  "), a.translator.T(i18n.RemoveLocalOnlyWarning))
	if !yes {
		confirmed, err := prompt(bufio.NewReader(a.in), a.out, a.translator.T(i18n.ContinueRemove), "N")
		if err != nil {
			return err
		}
		if !isYes(confirmed) {
			fmt.Fprintln(a.out, a.translator.T(i18n.RemoveCancelled))
			return nil
		}
	}

	targetID := target.ID()
	if cfg.EnvironmentCount() == 1 {
		backup, err := config.Archive(path, time.Now())
		if err != nil {
			return err
		}
		archived, err := config.Load(backup)
		if err != nil {
			return err
		}
		if err := archived.Validate(); err != nil {
			return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
		}
		fmt.Fprintf(a.out, "\n✓ %s\n✓ %s\n\n  %s\n  %s\n", a.translator.T(i18n.TargetRemoved), a.translator.T(i18n.ConfigArchived), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldIdentifier), targetID), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldBackup), backup))
		return nil
	}

	delete(project.Environments, target.EnvironmentID)
	if len(project.Environments) == 0 {
		delete(cfg.Projects, target.ProjectID)
	} else {
		cfg.Projects[target.ProjectID] = project
	}
	if err := cfg.Validate(); err != nil {
		return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	if err := config.Save(path, cfg); err != nil {
		return err
	}
	saved, err := config.Load(path)
	if err != nil {
		return err
	}
	if err := saved.Validate(); err != nil {
		return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	fmt.Fprintf(a.out, "\n✓ %s\n✓ %s\n\n  %s\n  %s\n", a.translator.T(i18n.TargetRemoved), a.translator.T(i18n.ConfigValid), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldIdentifier), targetID), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldPath), path))
	return nil
}
