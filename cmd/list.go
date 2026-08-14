package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/presentation"
)

func (a *app) newListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: a.translator.T(i18n.ListShort),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			cfg, _, err := a.loadConfig()
			if err != nil {
				return err
			}
			if err := cfg.Validate(); err != nil {
				return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
			}
			_, err = fmt.Fprint(a.out, presentation.RenderProjects(cfg, a.colorEnabled(), a.translator))
			return err
		},
	}
}
