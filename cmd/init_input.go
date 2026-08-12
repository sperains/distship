package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
)

func requiredPrompt(reader *bufio.Reader, out io.Writer, translator i18n.Translator, labelKey, hintKey i18n.Key) (string, error) {
	label := translator.T(labelKey)
	hint := translator.T(hintKey)
	fmt.Fprintf(out, "%s (%s): ", label, hint)
	value, err := readInput(reader)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", i18n.NewError(i18n.RequiredValue, label)
	}
	return value, nil
}

func prompt(reader *bufio.Reader, out io.Writer, label, defaultValue string) (string, error) {
	fmt.Fprintf(out, "%s [%s]: ", label, defaultValue)
	value, err := readInput(reader)
	if err != nil {
		return "", err
	}
	if value == "" {
		return defaultValue, nil
	}
	return value, nil
}

func optionalPrompt(reader *bufio.Reader, out io.Writer, label, hint string) (string, error) {
	fmt.Fprintf(out, "%s (%s): ", label, hint)
	return readInput(reader)
}

func readInput(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimSpace(line), nil
}

func isYes(value string) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "y", "yes", "是":
		return true
	default:
		return false
	}
}

func splitCommaList(value string) []string {
	if value == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if item := strings.TrimSpace(part); item != "" {
			result = append(result, item)
		}
	}
	return result
}

func expandHome(path string) (string, error) {
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", i18n.Wrap(i18n.ErrResolveLocalDirectory, err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, strings.TrimPrefix(path, "~/")), nil
	}
	return path, nil
}

func absoluteLocalDirectory(path string) (string, error) {
	expanded, err := expandHome(path)
	if err != nil {
		return "", err
	}
	absolute, err := filepath.Abs(expanded)
	if err != nil {
		return "", i18n.Wrap(i18n.ErrResolveLocalDirectory, err)
	}
	return filepath.Clean(absolute), nil
}

func askDeployTarget(reader *bufio.Reader, out io.Writer, translator i18n.Translator, defaultHost, defaultDirectory string) (string, string, error) {
	var value string
	var err error
	if defaultHost == "" {
		value, err = requiredPrompt(reader, out, translator, i18n.PromptSSHHost, i18n.HintSSHHost)
	} else {
		value, err = prompt(reader, out, translator.T(i18n.PromptSSHHost), defaultHost)
	}
	if err != nil {
		return "", "", err
	}
	host, directory, complete, err := parseDeployTarget(value)
	if err != nil {
		return "", "", err
	}
	if !complete {
		if defaultDirectory == "" {
			directory, err = requiredPrompt(reader, out, translator, i18n.PromptRemoteDirectory, i18n.HintAbsolutePath)
		} else {
			directory, err = prompt(reader, out, translator.T(i18n.PromptRemoteDirectory), defaultDirectory)
		}
		if err != nil {
			return "", "", err
		}
	}
	if !strings.HasPrefix(directory, "/") {
		return "", "", i18n.NewError(i18n.InvalidDeployTarget)
	}
	return host, directory, nil
}

func parseDeployTarget(value string) (string, string, bool, error) {
	value = strings.TrimSpace(value)
	separator := strings.Index(value, ":/")
	host := value
	directory := ""
	complete := false
	if separator >= 0 {
		host = strings.TrimSpace(value[:separator])
		directory = strings.TrimSpace(value[separator+1:])
		complete = true
	} else if strings.Contains(value, "/") {
		return "", "", false, i18n.NewError(i18n.InvalidDeployTarget)
	}
	if err := validateSSHTarget(host); err != nil {
		return "", "", false, err
	}
	if complete && !strings.HasPrefix(directory, "/") {
		return "", "", false, i18n.NewError(i18n.InvalidDeployTarget)
	}
	return host, directory, complete, nil
}

func validateSSHTarget(value string) error {
	if !config.IsValidSSHTarget(value) {
		return i18n.NewError(i18n.InvalidSSHTarget)
	}
	return nil
}

func splitCommand(input string) ([]string, error) {
	var result []string
	var current strings.Builder
	var quote rune
	escaped := false
	flush := func() {
		if current.Len() > 0 {
			result = append(result, current.String())
			current.Reset()
		}
	}
	for _, char := range input {
		if escaped {
			current.WriteRune(char)
			escaped = false
			continue
		}
		if char == '\\' && quote != '\'' {
			escaped = true
			continue
		}
		if quote != 0 {
			if char == quote {
				quote = 0
			} else {
				current.WriteRune(char)
			}
			continue
		}
		switch char {
		case '\'', '"':
			quote = char
		case ' ', '\t':
			flush()
		default:
			current.WriteRune(char)
		}
	}
	if escaped || quote != 0 {
		return nil, i18n.NewError(i18n.UnclosedCommand)
	}
	flush()
	if len(result) == 0 {
		return nil, i18n.NewError(i18n.EmptyCommand)
	}
	return result, nil
}
