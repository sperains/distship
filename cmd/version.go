package cmd

import (
	"runtime/debug"
	"strings"
)

type buildMetadata struct {
	version string
	commit  string
	date    string
}

func currentBuildMetadata() buildMetadata {
	info, _ := debug.ReadBuildInfo()
	return resolveBuildMetadata(version, commit, date, info)
}

func resolveBuildMetadata(linkedVersion, linkedCommit, linkedDate string, info *debug.BuildInfo) buildMetadata {
	metadata := buildMetadata{version: linkedVersion, commit: linkedCommit, date: linkedDate}
	if info == nil {
		return metadata
	}

	if metadata.version == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		metadata.version = info.Main.Version
	}

	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			if metadata.commit == "" || metadata.commit == "none" {
				metadata.commit = setting.Value
			}
		case "vcs.time":
			if metadata.date == "" || metadata.date == "unknown" || strings.HasPrefix(metadata.date, "0001-01-01") {
				metadata.date = setting.Value
			}
		}
	}

	return metadata
}
