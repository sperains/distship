package cmd

import (
	"fmt"
	"os"

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
			color := !a.noColor && os.Getenv("NO_COLOR") == "" && isTerminal(a.out)
			_, err = fmt.Fprint(a.out, presentation.RenderProjects(cfg, color, a.translator))
			return err
		},
	}
}

func isTerminal(w any) bool {
	file, ok := w.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}
