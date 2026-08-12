package cmd

import (
	"strings"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
)

func (a *app) resolveTarget(targetID string) (config.TargetRef, config.Environment, error) {
	ref, ok := config.ParseTargetID(targetID)
	if !ok {
		return config.TargetRef{}, config.Environment{}, i18n.NewError(i18n.InvalidTargetID)
	}
	cfg, _, err := a.loadConfig()
	if err != nil {
		return config.TargetRef{}, config.Environment{}, err
	}
	if err := cfg.Validate(); err != nil {
		return config.TargetRef{}, config.Environment{}, i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	_, environment, exists := cfg.Target(ref)
	if !exists {
		return config.TargetRef{}, config.Environment{}, i18n.NewError(i18n.TargetNotFound, ref.ID(), strings.Join(cfg.TargetIDs(), "\n  "))
	}
	return ref, environment, nil
}
