package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/history"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/presentation"
)

func (a *app) newHistoryCommand() *cobra.Command {
	const defaultLimit = 10
	limit := defaultLimit
	command := &cobra.Command{
		Use:   "history [target-id]",
		Short: a.translator.T(i18n.HistoryShort),
		Args: func(_ *cobra.Command, args []string) error {
			if len(args) > 1 {
				return i18n.NewError(i18n.InvalidHistoryArguments)
			}
			return nil
		},
		RunE: func(_ *cobra.Command, args []string) error {
			if limit <= 0 {
				return i18n.NewError(i18n.InvalidHistoryLimit)
			}
			var ref *config.TargetRef
			targetID := ""
			if len(args) == 1 {
				parsed, valid := config.ParseTargetID(args[0])
				if !valid {
					return i18n.NewError(i18n.InvalidTargetID)
				}
				ref = &parsed
				targetID = parsed.ID()
			}
			path, err := history.DefaultPath()
			if err != nil {
				return i18n.Wrap(i18n.HistoryReadFailed, err)
			}
			records, err := history.List(path, ref, limit)
			if err != nil {
				return i18n.Wrap(i18n.HistoryReadFailed, err)
			}
			_, err = fmt.Fprint(a.out, presentation.RenderHistory(records, targetID, a.colorEnabled(), a.translator))
			return err
		},
	}
	command.Flags().IntVar(&limit, "limit", defaultLimit, a.translator.T(i18n.FlagHistoryLimit))
	return command
}
