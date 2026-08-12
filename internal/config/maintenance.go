package config

import (
	"os"
	"strconv"
	"time"

	"github.com/sperains/distship/internal/i18n"
)

func Archive(path string, archivedAt time.Time) (string, error) {
	base := path + ".bak-" + archivedAt.Format("20060102T150405")
	backup := base
	for suffix := 1; ; suffix++ {
		_, err := os.Stat(backup)
		if os.IsNotExist(err) {
			break
		}
		if err != nil {
			return "", i18n.Wrap(i18n.ErrArchiveConfig, err, path)
		}
		backup = base + "-" + strconv.Itoa(suffix)
	}
	if err := os.Rename(path, backup); err != nil {
		return "", i18n.Wrap(i18n.ErrArchiveConfig, err, path)
	}
	return backup, nil
}
