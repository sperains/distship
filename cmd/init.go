package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/sperains/distship/internal/config"
	"github.com/sperains/distship/internal/i18n"
	projectinfo "github.com/sperains/distship/internal/project"
)

type initAnswers struct {
	projectID       string
	projectName     string
	environmentID   string
	environmentName string
	directory       string
	build           []string
	artifact        string
	host            string
	remoteDirectory string
	allowedBranches []string
	dirty           string
}

func (a *app) newInitCommand() *cobra.Command {
	var advanced bool
	command := &cobra.Command{
		Use:   "init [directory]",
		Short: a.translator.T(i18n.InitShort),
		Args:  cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			directory := ""
			if len(args) == 1 {
				directory = args[0]
			}
			return a.runInit(directory, advanced)
		},
	}
	command.Flags().BoolVar(&advanced, "advanced", false, a.translator.T(i18n.FlagAdvanced))
	return command
}

func (a *app) runInit(directoryArgument string, advanced bool) error {
	path, err := config.InitPath(a.configPath)
	if err != nil {
		return err
	}
	cfg, err := loadForInit(path)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(a.in)
	fmt.Fprintln(a.out, a.translator.T(i18n.InitTitle))
	fmt.Fprintln(a.out)
	if directoryArgument == "" {
		fmt.Fprintln(a.out, a.translator.T(i18n.InitQuickHint))
		fmt.Fprintln(a.out)
	}
	directory, err := resolveInitDirectory(reader, a.out, a.translator, cfg, directoryArgument)
	if err != nil {
		return err
	}

	var answers initAnswers
	if advanced {
		answers, err = askAdvancedInitQuestions(reader, a.out, a.translator, cfg, directory)
	} else {
		answers, err = askQuickInitQuestions(reader, a.out, a.translator, cfg, directory)
	}
	if err != nil {
		return err
	}

	currentProject := cfg.Projects[answers.projectID]
	currentEnvironment, targetExists := currentProject.Environments[answers.environmentID]
	candidate := environmentFromAnswers(answers)
	candidateConfig := cfg.Clone()
	candidateProject := candidateConfig.Projects[answers.projectID]
	if candidateProject.Environments == nil {
		candidateProject.Environments = make(map[string]config.Environment)
	}
	candidateProject.Name = answers.projectName
	candidateProject.Environments[answers.environmentID] = candidate
	candidateConfig.Projects[answers.projectID] = candidateProject
	if err := candidateConfig.Validate(); err != nil {
		return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	if targetExists {
		fmt.Fprintln(a.out, a.translator.T(i18n.ExistingTargetFound, config.TargetID(answers.projectID, answers.environmentID)))
		if !printInitChanges(a.out, currentProject.Name, currentEnvironment, answers, candidate, a.translator) {
			fmt.Fprintln(a.out, a.translator.T(i18n.NoChanges))
			return nil
		}
	} else {
		printInitPreview(a.out, path, answers, a.translator)
	}
	fmt.Fprintln(a.out, a.translator.T(i18n.InitLocalOnlyWarning))

	confirmation := i18n.ConfirmCreate
	if targetExists {
		confirmation = i18n.ConfirmUpdate
	}
	confirmed, err := prompt(reader, a.out, a.translator.T(confirmation), "N")
	if err != nil {
		return err
	}
	if !isYes(confirmed) {
		fmt.Fprintln(a.out, a.translator.T(i18n.Cancelled))
		return nil
	}

	if err := config.Save(path, candidateConfig); err != nil {
		return err
	}
	saved, err := config.Load(path)
	if err != nil {
		return err
	}
	if err := saved.Validate(); err != nil {
		return i18n.NewError(i18n.ConfigInvalid, a.translator.Error(err))
	}
	resultMessage := i18n.ConfigWritten
	if targetExists {
		resultMessage = i18n.ConfigUpdated
	}
	fmt.Fprintf(a.out, "\n✓ %s\n✓ %s\n\n  %s\n  %s\n\n%s\n", a.translator.T(resultMessage), a.translator.T(i18n.ConfigValid), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldPath), path), a.translator.T(i18n.FieldLine, a.translator.T(i18n.FieldIdentifier), config.TargetID(answers.projectID, answers.environmentID)), a.translator.T(i18n.NextStep))
	return nil
}

func loadForInit(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err == nil {
		return cfg, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return config.New(), nil
	}
	return nil, err
}

func resolveInitDirectory(reader *bufio.Reader, out io.Writer, translator i18n.Translator, cfg *config.Config, argument string) (string, error) {
	directory := argument
	var err error
	if directory == "" {
		currentDirectory, currentErr := os.Getwd()
		if currentErr == nil {
			currentDirectory, currentErr = absoluteLocalDirectory(currentDirectory)
		}
		if currentErr == nil && (cfg.HasTargetDirectory(currentDirectory) || len(projectinfo.Discover(currentDirectory, "test").Build) > 0) {
			directory = currentDirectory
			fmt.Fprintf(out, "✓ %s\n  %s\n\n", translator.T(i18n.CurrentDirectoryDetected), directory)
		} else {
			directory, err = requiredPrompt(reader, out, translator, i18n.PromptLocalDirectory, i18n.HintLocalDirectory)
			if err != nil {
				return "", err
			}
		}
	}
	directory, err = absoluteLocalDirectory(directory)
	if err != nil {
		return "", err
	}
	info, statErr := os.Stat(directory)
	if os.IsNotExist(statErr) {
		return "", i18n.NewError(i18n.LocalDirectoryNotFound, directory)
	}
	if statErr != nil {
		return "", i18n.Wrap(i18n.ErrCheckLocalDirectory, statErr, directory)
	}
	if !info.IsDir() {
		return "", i18n.NewError(i18n.LocalPathNotDirectory, directory)
	}
	return directory, nil
}

func printProjectDetection(out io.Writer, directory string, defaults projectinfo.Defaults, translator i18n.Translator) {
	value := func(value string) string {
		if value == "" {
			return translator.T(i18n.NotDetected)
		}
		return value
	}
	lines := []string{
		translator.T(i18n.FieldLine, translator.T(i18n.FieldLocalDirectory), directory),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldProject), value(defaults.ProjectName)),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldProjectType), value(defaults.ProjectType)),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldPackageManager), value(defaults.PackageManager)),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldGitBranch), value(defaults.Branch)),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldBuildCommand), value(strings.Join(defaults.Build, " "))),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldArtifact), value(defaults.Artifact)),
	}
	fmt.Fprintf(out, "%s\n\n  %s\n\n", translator.T(i18n.DetectionTitle), strings.Join(lines, "\n  "))
}

func environmentFromAnswers(answers initAnswers) config.Environment {
	return config.Environment{
		Name:      answers.environmentName,
		Directory: answers.directory,
		Build:     append([]string(nil), answers.build...),
		Artifact:  answers.artifact,
		Target: config.Target{
			Host:      answers.host,
			Directory: answers.remoteDirectory,
		},
		Git: config.GitPolicy{
			AllowedBranches: append([]string(nil), answers.allowedBranches...),
			Dirty:           answers.dirty,
		},
	}
}

func printInitChanges(out io.Writer, currentProjectName string, current config.Environment, answers initAnswers, candidate config.Environment, translator i18n.Translator) bool {
	type change struct {
		label   i18n.Key
		current string
		new     string
	}
	changes := []change{
		{i18n.FieldProject, currentProjectName, answers.projectName},
		{i18n.FieldEnvironment, current.Name, candidate.Name},
		{i18n.FieldLocalDirectory, current.Directory, candidate.Directory},
		{i18n.FieldBuildCommand, strings.Join(current.Build, " "), strings.Join(candidate.Build, " ")},
		{i18n.FieldArtifact, current.Artifact, candidate.Artifact},
		{i18n.FieldDeployTarget, current.Target.String(), candidate.Target.String()},
		{i18n.FieldAllowedBranches, formatBranches(current.Git.AllowedBranches, translator), formatBranches(candidate.Git.AllowedBranches, translator)},
		{i18n.FieldWorkingTree, current.Git.Dirty, candidate.Git.Dirty},
	}
	var output strings.Builder
	count := 0
	for _, item := range changes {
		if item.current == item.new {
			continue
		}
		fmt.Fprintf(&output, "  %s\n    %s\n    %s\n", translator.T(item.label), translator.T(i18n.FieldLine, translator.T(i18n.FieldCurrent), item.current), translator.T(i18n.FieldLine, translator.T(i18n.FieldNew), item.new))
		count++
	}
	if count == 0 {
		return false
	}
	fmt.Fprintf(out, "\n%s\n\n%s\n", translator.T(i18n.ChangesTitle), output.String())
	return true
}

func formatBranches(branches []string, translator i18n.Translator) string {
	if len(branches) == 0 {
		return translator.T(i18n.AnyBranchWarning)
	}
	return strings.Join(branches, ", ")
}

func askQuickInitQuestions(reader *bufio.Reader, out io.Writer, translator i18n.Translator, cfg *config.Config, directory string) (initAnswers, error) {
	answers := initAnswers{directory: directory}
	defaults := projectinfo.Discover(directory, "test")
	printProjectDetection(out, directory, defaults, translator)
	var err error
	projectDefault, environmentDefault := initTargetDefaults(cfg, directory, defaults)
	answers.projectID, err = prompt(reader, out, translator.T(i18n.PromptProjectID), projectDefault)
	if err != nil {
		return answers, err
	}
	answers.environmentID, err = prompt(reader, out, translator.T(i18n.PromptEnvironmentID), environmentDefault)
	if err != nil {
		return answers, err
	}
	if answers.environmentID != "test" {
		defaults = projectinfo.Discover(answers.directory, answers.environmentID)
	}

	answers.projectName = defaults.ProjectName
	if answers.projectName == "" || answers.projectID != defaults.ProjectID {
		answers.projectName = answers.projectID
	}
	answers.environmentName = answers.environmentID
	existingProject, projectExists := cfg.Projects[answers.projectID]
	existingEnvironment, targetExists := existingProject.Environments[answers.environmentID]
	if projectExists {
		answers.projectName = existingProject.Name
		if targetExists {
			answers.environmentName = existingEnvironment.Name
		}
	}
	answers.artifact = defaults.Artifact
	answers.dirty = "warn"
	if targetExists {
		answers.artifact = existingEnvironment.Artifact
		answers.dirty = existingEnvironment.Git.Dirty
		answers.allowedBranches = append([]string(nil), existingEnvironment.Git.AllowedBranches...)
	} else if defaults.Branch != "" {
		answers.allowedBranches = []string{defaults.Branch}
	}

	buildDefault := strings.Join(defaults.Build, " ")
	if targetExists {
		buildDefault = strings.Join(existingEnvironment.Build, " ")
	}
	var buildText string
	if buildDefault == "" {
		buildText, err = requiredPrompt(reader, out, translator, i18n.PromptBuildCommand, i18n.HintBuildCommand)
	} else {
		buildText, err = prompt(reader, out, translator.T(i18n.PromptBuildCommand), buildDefault)
	}
	if err != nil {
		return answers, err
	}
	answers.build, err = splitCommand(buildText)
	if err != nil {
		return answers, i18n.NewError(i18n.InvalidBuildCommand, translator.Error(err))
	}

	if answers.artifact == "" {
		answers.artifact, err = requiredPrompt(reader, out, translator, i18n.PromptArtifact, i18n.HintArtifact)
		if err != nil {
			return answers, err
		}
	}

	defaultHost := ""
	defaultRemoteDirectory := ""
	if targetExists {
		defaultHost = existingEnvironment.Target.Host
		defaultRemoteDirectory = existingEnvironment.Target.Directory
	}
	answers.host, answers.remoteDirectory, err = askDeployTarget(reader, out, translator, defaultHost, defaultRemoteDirectory)
	if err != nil {
		return answers, err
	}
	return answers, nil
}

func askAdvancedInitQuestions(reader *bufio.Reader, out io.Writer, translator i18n.Translator, cfg *config.Config, directory string) (initAnswers, error) {
	answers := initAnswers{directory: directory}
	defaults := projectinfo.Discover(directory, "test")
	printProjectDetection(out, directory, defaults, translator)
	var err error
	projectDefault, environmentDefault := initTargetDefaults(cfg, directory, defaults)
	if answers.projectID, err = prompt(reader, out, translator.T(i18n.PromptProjectID), projectDefault); err != nil {
		return answers, err
	}
	projectNameDefault := defaults.ProjectName
	if projectNameDefault == "" || answers.projectID != defaults.ProjectID {
		projectNameDefault = answers.projectID
	}
	if project, exists := cfg.Projects[answers.projectID]; exists {
		projectNameDefault = project.Name
	}
	if answers.projectName, err = prompt(reader, out, translator.T(i18n.PromptProjectName), projectNameDefault); err != nil {
		return answers, err
	}
	if answers.environmentID, err = prompt(reader, out, translator.T(i18n.PromptEnvironmentID), environmentDefault); err != nil {
		return answers, err
	}
	if answers.environmentID != "test" {
		defaults = projectinfo.Discover(directory, answers.environmentID)
	}
	existingEnvironment, targetExists := cfg.Projects[answers.projectID].Environments[answers.environmentID]
	environmentNameDefault := answers.environmentID
	if targetExists {
		environmentNameDefault = existingEnvironment.Name
	}
	if answers.environmentName, err = prompt(reader, out, translator.T(i18n.PromptEnvironmentName), environmentNameDefault); err != nil {
		return answers, err
	}
	buildDefault := strings.Join(defaults.Build, " ")
	if targetExists {
		buildDefault = strings.Join(existingEnvironment.Build, " ")
	}
	var buildText string
	if buildDefault == "" {
		buildText, err = requiredPrompt(reader, out, translator, i18n.PromptBuildCommand, i18n.HintBuildCommand)
	} else {
		buildText, err = prompt(reader, out, translator.T(i18n.PromptBuildCommand), buildDefault)
	}
	if err != nil {
		return answers, err
	}
	answers.build, err = splitCommand(buildText)
	if err != nil {
		return answers, i18n.NewError(i18n.InvalidBuildCommand, translator.Error(err))
	}
	artifactDefault := defaults.Artifact
	if targetExists {
		artifactDefault = existingEnvironment.Artifact
	}
	if artifactDefault == "" {
		answers.artifact, err = requiredPrompt(reader, out, translator, i18n.PromptArtifact, i18n.HintArtifact)
	} else {
		answers.artifact, err = prompt(reader, out, translator.T(i18n.PromptArtifact), artifactDefault)
	}
	if err != nil {
		return answers, err
	}
	if targetExists {
		if answers.host, err = prompt(reader, out, translator.T(i18n.PromptSSHHost), existingEnvironment.Target.Host); err != nil {
			return answers, err
		}
		if answers.remoteDirectory, err = prompt(reader, out, translator.T(i18n.PromptRemoteDirectory), existingEnvironment.Target.Directory); err != nil {
			return answers, err
		}
	} else {
		if answers.host, err = requiredPrompt(reader, out, translator, i18n.PromptSSHHost, i18n.HintSSHHost); err != nil {
			return answers, err
		}
		if answers.remoteDirectory, err = requiredPrompt(reader, out, translator, i18n.PromptRemoteDirectory, i18n.HintAbsolutePath); err != nil {
			return answers, err
		}
	}
	if err := validateSSHTarget(answers.host); err != nil {
		return answers, err
	}
	anyBranchHint := translator.T(i18n.HintAnyBranch)
	branchDefault := defaults.Branch
	if targetExists {
		branchDefault = strings.Join(existingEnvironment.Git.AllowedBranches, ",")
	}
	var branches string
	if branchDefault == "" {
		branches, err = optionalPrompt(reader, out, translator.T(i18n.PromptAllowedBranches), anyBranchHint)
	} else {
		branches, err = prompt(reader, out, translator.T(i18n.PromptAllowedBranches), branchDefault)
	}
	if err != nil {
		return answers, err
	}
	answers.allowedBranches = splitCommaList(branches)
	dirtyDefault := "warn"
	if targetExists {
		dirtyDefault = existingEnvironment.Git.Dirty
	}
	if answers.dirty, err = prompt(reader, out, translator.T(i18n.PromptDirtyPolicy), dirtyDefault); err != nil {
		return answers, err
	}
	return answers, nil
}

func initTargetDefaults(cfg *config.Config, directory string, defaults projectinfo.Defaults) (string, string) {
	projectID := defaults.ProjectID
	if projectID == "" {
		projectID = "project"
	}
	environmentID := "test"
	preferences := []string{environmentID}
	if defaults.Branch != "" && defaults.Branch != environmentID {
		preferences = append([]string{defaults.Branch}, preferences...)
	}
	if target, found := cfg.FindTargetByDirectory(directory, preferences...); found {
		return target.ProjectID, target.EnvironmentID
	}
	return projectID, environmentID
}

func printInitPreview(out io.Writer, path string, a initAnswers, translator i18n.Translator) {
	branches := translator.T(i18n.AnyBranchWarning)
	if len(a.allowedBranches) > 0 {
		branches = strings.Join(a.allowedBranches, ", ")
	}
	lines := []string{
		translator.T(i18n.FieldLine, translator.T(i18n.FieldProject), a.projectID+" · "+a.projectName),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldEnvironment), a.environmentID+" · "+a.environmentName),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldLocalDirectory), a.directory),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldBuildCommand), strings.Join(a.build, " ")),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldArtifact), a.artifact),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldDeployTarget), (config.Target{Host: a.host, Directory: a.remoteDirectory}).String()),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldAllowedBranches), branches),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldWorkingTree), a.dirty),
		translator.T(i18n.FieldLine, translator.T(i18n.FieldConfigFile), path),
	}
	fmt.Fprintf(out, "\n%s\n\n  %s\n\n", translator.T(i18n.PreviewTitle), strings.Join(lines, "\n  "))
}
