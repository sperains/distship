package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
	"github.com/sperains/distship/internal/process"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type app struct {
	in         io.Reader
	out        io.Writer
	configPath string
	noColor    bool
	language   string
	translator i18n.Translator
	runner     process.Runner
}

func Execute() error {
	root, application := buildRootCommand(os.Stdin, os.Stdout, os.Stderr, languageFlag(os.Args[1:]))
	if err := root.Execute(); err != nil {
		return errors.New(application.translator.Error(err))
	}
	return nil
}

func newRootCommand(in io.Reader, out, errOut io.Writer) *cobra.Command {
	root, _ := buildRootCommand(in, out, errOut, "")
	return root
}

func buildRootCommand(in io.Reader, out, errOut io.Writer, initialLanguage string) (*cobra.Command, *app) {
	translator, _ := i18n.Resolve(initialLanguage)
	a := &app{in: in, out: out, language: "auto", translator: translator, runner: process.OSRunner{}}
	root := &cobra.Command{
		Use:           "distship",
		Short:         translator.T(i18n.RootShort),
		SilenceUsage:  true,
		SilenceErrors: true,
		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			resolved, err := i18n.Resolve(a.language)
			if err != nil {
				return err
			}
			a.translator = resolved
			return nil
		},
		RunE: func(cmd *cobra.Command, _ []string) error {
			return cmd.Help()
		},
	}
	root.SetIn(in)
	root.SetOut(out)
	root.SetErr(errOut)
	root.CompletionOptions.DisableDefaultCmd = true
	root.SetHelpTemplate(helpTemplate(translator))
	root.PersistentFlags().StringVar(&a.configPath, "config", "", translator.T(i18n.FlagConfig))
	root.PersistentFlags().BoolVar(&a.noColor, "no-color", false, translator.T(i18n.FlagNoColor))
	root.PersistentFlags().StringVar(&a.language, "lang", "auto", translator.T(i18n.FlagLanguage))
	root.PersistentFlags().Lookup("lang").DefValue = ""
	root.AddCommand(
		a.newInitCommand(),
		a.newListCommand(),
		a.newCheckCommand(),
		a.newDeployCommand(),
		a.newConfigCommand(),
		a.newVersionCommand(),
	)
	localizeHelp(root, translator)
	return root, a
}

func (a *app) loadConfig() (*config.Config, string, error) {
	path, err := config.FindPath(a.configPath)
	if err != nil {
		return nil, "", err
	}
	cfg, err := config.Load(path)
	if err != nil {
		return nil, "", err
	}
	return cfg, path, nil
}

func (a *app) newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: a.translator.T(i18n.VersionShort),
		Args:  cobra.NoArgs,
		RunE: func(_ *cobra.Command, _ []string) error {
			_, err := fmt.Fprintf(a.out, "distship %s\ncommit: %s\nbuilt: %s\n", version, commit, date)
			return err
		},
	}
}

func languageFlag(args []string) string {
	for index, argument := range args {
		if strings.HasPrefix(argument, "--lang=") {
			return strings.TrimPrefix(argument, "--lang=")
		}
		if argument == "--lang" && index+1 < len(args) {
			return args[index+1]
		}
	}
	return ""
}

func helpTemplate(translator i18n.Translator) string {
	return fmt.Sprintf(`{{with (or .Long .Short)}}{{. | trimTrailingWhitespaces}}

{{end}}{{if or .Runnable .HasSubCommands}}%s
  {{if .HasAvailableSubCommands}}{{.CommandPath}} [command]{{else}}{{.UseLine}}{{end}}{{end}}{{if gt (len .Aliases) 0}}

%s
  {{.NameAndAliases}}{{end}}{{if .HasExample}}

%s
{{.Example}}{{end}}{{if .HasAvailableSubCommands}}

%s{{range .Commands}}{{if (or .IsAvailableCommand (eq .Name "help"))}}
  {{rpad .Name .NamePadding }} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableLocalFlags}}

%s
{{.LocalFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasAvailableInheritedFlags}}

%s
{{.InheritedFlags.FlagUsages | trimTrailingWhitespaces}}{{end}}{{if .HasHelpSubCommands}}

%s{{range .Commands}}{{if .IsAdditionalHelpTopicCommand}}
  {{rpad .CommandPath .CommandPathPadding}} {{.Short}}{{end}}{{end}}{{end}}{{if .HasAvailableSubCommands}}

%s
{{end}}`, translator.T(i18n.HelpUsage), translator.T(i18n.HelpAliases), translator.T(i18n.HelpExamples), translator.T(i18n.HelpAvailableCommands), translator.T(i18n.HelpFlags), translator.T(i18n.HelpGlobalFlags), translator.T(i18n.HelpAdditionalTopics), translator.T(i18n.HelpMoreInformation, "{{.CommandPath}}"))
}

func localizeHelp(root *cobra.Command, translator i18n.Translator) {
	root.InitDefaultHelpCmd()
	var visit func(*cobra.Command)
	visit = func(command *cobra.Command) {
		command.InitDefaultHelpFlag()
		if helpFlag := command.Flags().Lookup("help"); helpFlag != nil {
			helpFlag.Usage = translator.T(i18n.FlagHelp, command.Name())
		}
		for _, child := range command.Commands() {
			if child.Name() == "help" {
				child.Short = translator.T(i18n.HelpCommandShort)
			}
			visit(child)
		}
	}
	visit(root)
}
