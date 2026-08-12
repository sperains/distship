package i18n

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

type Language string

const (
	English           Language = "en"
	SimplifiedChinese Language = "zh-CN"
)

type Key string

const (
	RootShort                     Key = "root.short"
	FlagConfig                    Key = "flag.config"
	FlagNoColor                   Key = "flag.no_color"
	FlagLanguage                  Key = "flag.language"
	FlagAdvanced                  Key = "flag.advanced"
	FlagYes                       Key = "flag.yes"
	FlagHelp                      Key = "flag.help"
	HelpCommandShort              Key = "help.command_short"
	HelpUsage                     Key = "help.usage"
	HelpAliases                   Key = "help.aliases"
	HelpExamples                  Key = "help.examples"
	HelpAvailableCommands         Key = "help.available_commands"
	HelpAdditionalTopics          Key = "help.additional_topics"
	HelpFlags                     Key = "help.flags"
	HelpGlobalFlags               Key = "help.global_flags"
	HelpMoreInformation           Key = "help.more_information"
	VersionShort                  Key = "version.short"
	ConfigShort                   Key = "config.short"
	ConfigValidateShort           Key = "config.validate.short"
	ConfigInvalid                 Key = "config.invalid"
	ConfigValid                   Key = "config.valid"
	ConfigRemoveShort             Key = "config.remove.short"
	RemovePreviewTitle            Key = "config.remove.preview_title"
	RemoveLocalOnlyWarning        Key = "config.remove.local_only_warning"
	ContinueRemove                Key = "config.remove.continue"
	RemoveCancelled               Key = "config.remove.cancelled"
	TargetRemoved                 Key = "config.remove.target_removed"
	ConfigArchived                Key = "config.remove.config_archived"
	TargetNotFound                Key = "config.remove.target_not_found"
	InvalidTargetID               Key = "config.remove.invalid_target_id"
	ListShort                     Key = "list.short"
	CheckShort                    Key = "check.short"
	CheckTitle                    Key = "check.title"
	CheckReadOnlyNotice           Key = "check.read_only_notice"
	CheckPassed                   Key = "check.passed"
	CheckPassedWithWarnings       Key = "check.passed_with_warnings"
	CheckSummary                  Key = "check.summary"
	CheckRemoteCreatable          Key = "check.remote_creatable"
	WarningGitUnavailable         Key = "check.warning.git_unavailable"
	WarningNotGit                 Key = "check.warning.not_git"
	WarningDirtyWorkingTree       Key = "check.warning.dirty_working_tree"
	WarningDetachedHead           Key = "check.warning.detached_head"
	DetachedHead                  Key = "value.detached_head"
	WorkingTreeClean              Key = "value.working_tree_clean"
	WorkingTreeDirty              Key = "value.working_tree_dirty"
	RecentCommitsTitle            Key = "check.recent_commits"
	NoRecentCommits               Key = "check.no_recent_commits"
	DeployShort                   Key = "deploy.short"
	FlagDryRun                    Key = "deploy.flag.dry_run"
	DeployTitle                   Key = "deploy.title"
	DeployReady                   Key = "deploy.ready"
	DeployReadyWithWarnings       Key = "deploy.ready_with_warnings"
	DeployDryRun                  Key = "deploy.dry_run"
	DeployRecentCommits           Key = "deploy.recent_commits"
	DeployChangesSince            Key = "deploy.changes_since"
	DeployNoLocalHistory          Key = "deploy.no_local_history"
	DeployNoNewCommits            Key = "deploy.no_new_commits"
	DeployHistoryMissing          Key = "deploy.history_missing"
	DeployHistoryDiverged         Key = "deploy.history_diverged"
	DeployPreviousDirty           Key = "deploy.previous_dirty"
	DeployChecks                  Key = "deploy.checks"
	DeployDryRunChecks            Key = "deploy.dry_run_checks"
	DeployActionNotice            Key = "deploy.action_notice"
	DeployDryRunNotice            Key = "deploy.dry_run_notice"
	ConfirmDeploy                 Key = "deploy.confirm"
	DeployCancelled               Key = "deploy.cancelled"
	BuildTitle                    Key = "deploy.build_title"
	BuildCompleted                Key = "deploy.build_completed"
	UploadTitle                   Key = "deploy.upload_title"
	DeployCompleted               Key = "deploy.completed"
	DeployHistoryWritten          Key = "deploy.history_written"
	InitShort                     Key = "init.short"
	InitTitle                     Key = "init.title"
	InitQuickHint                 Key = "init.quick_hint"
	CurrentDirectoryDetected      Key = "init.current_directory_detected"
	DetectionTitle                Key = "init.detection_title"
	ExistingTargetFound           Key = "init.existing_target_found"
	ChangesTitle                  Key = "init.changes_title"
	NoChanges                     Key = "init.no_changes"
	ConfirmCreate                 Key = "init.confirm_create"
	ConfirmUpdate                 Key = "init.confirm_update"
	ConfigUpdated                 Key = "init.config_updated"
	InitLocalOnlyWarning          Key = "init.local_only_warning"
	Cancelled                     Key = "init.cancelled"
	ConfigWritten                 Key = "init.config_written"
	NextStep                      Key = "init.next_step"
	TargetSetupTitle              Key = "init.target_setup_title"
	TargetIDUsage                 Key = "init.target_id_usage"
	PromptTargetID                Key = "prompt.target_id"
	PromptProjectName             Key = "prompt.project_name"
	PromptEnvironment             Key = "prompt.deployment_environment"
	PromptEnvironmentName         Key = "prompt.environment_name"
	PromptLocalDirectory          Key = "prompt.local_directory"
	HintLocalDirectory            Key = "hint.local_directory"
	HintAbsolutePath              Key = "hint.absolute_path"
	PromptBuildCommand            Key = "prompt.build_command"
	HintBuildCommand              Key = "hint.build_command"
	PromptArtifact                Key = "prompt.artifact"
	PromptSSHHost                 Key = "prompt.ssh_host"
	HintSSHHost                   Key = "hint.ssh_host"
	PromptRemoteDirectory         Key = "prompt.remote_directory"
	PromptAllowedBranches         Key = "prompt.allowed_branches"
	HintAnyBranch                 Key = "hint.any_branch"
	PromptDirtyPolicy             Key = "prompt.dirty_policy"
	AnyBranchWarning              Key = "value.any_branch_warning"
	NotDetected                   Key = "value.not_detected"
	HintArtifact                  Key = "hint.artifact"
	PreviewTitle                  Key = "preview.title"
	FieldProject                  Key = "field.project"
	FieldEnvironment              Key = "field.environment"
	FieldLocalDirectory           Key = "field.local_directory"
	FieldBuildCommand             Key = "field.build_command"
	FieldArtifact                 Key = "field.artifact"
	FieldDeployTarget             Key = "field.deploy_target"
	FieldAllowedBranches          Key = "field.allowed_branches"
	FieldWorkingTree              Key = "field.working_tree"
	FieldConfigFile               Key = "field.config_file"
	FieldPath                     Key = "field.path"
	FieldProjects                 Key = "field.projects"
	FieldTargets                  Key = "field.targets"
	FieldIdentifier               Key = "field.identifier"
	FieldTargetID                 Key = "field.target_id"
	FieldLocal                    Key = "field.local"
	FieldRemote                   Key = "field.remote"
	FieldBranches                 Key = "field.branches"
	FieldProjectType              Key = "field.project_type"
	FieldPackageManager           Key = "field.package_manager"
	FieldGitBranch                Key = "field.git_branch"
	FieldSource                   Key = "field.source"
	FieldBuild                    Key = "field.build"
	FieldChecks                   Key = "field.checks"
	FieldWarnings                 Key = "field.warnings"
	FieldDuration                 Key = "field.duration"
	FieldCurrent                  Key = "field.current"
	FieldNew                      Key = "field.new"
	FieldBackup                   Key = "field.backup"
	FieldLine                     Key = "field.line"
	RequiredValue                 Key = "error.required_value"
	InvalidBuildCommand           Key = "error.invalid_build_command"
	InvalidEnvironmentID          Key = "error.invalid_deployment_environment"
	InvalidDeployTarget           Key = "error.invalid_deploy_target"
	InvalidSSHTarget              Key = "error.invalid_ssh_target"
	CheckTimedOut                 Key = "error.check_timed_out"
	CheckLocalDirectoryFailed     Key = "error.preflight_local_directory"
	CheckBuildToolFailed          Key = "error.preflight_build_tool"
	CheckGitFailed                Key = "error.preflight_git"
	CheckBranchFailed             Key = "error.preflight_branch"
	CheckDirtyDenied              Key = "error.preflight_dirty"
	CheckSSHFailed                Key = "error.preflight_ssh"
	CheckRemoteFailed             Key = "error.preflight_remote"
	DeployRsyncUnavailable        Key = "error.deploy_rsync_unavailable"
	DeployHistoryFailed           Key = "error.deploy_history"
	DeployBuildFailed             Key = "error.deploy_build"
	DeployRemoteCreateFailed      Key = "error.deploy_remote_create"
	DeployUploadFailed            Key = "error.deploy_upload"
	DeployLockFailed              Key = "error.deploy_lock"
	DeployLockReleaseFailed       Key = "error.deploy_lock_release"
	DeployHistoryWriteFailed      Key = "error.deploy_history_write"
	LocalDirectoryNotFound        Key = "error.local_directory_not_found"
	LocalPathNotDirectory         Key = "error.local_path_not_directory"
	UnclosedCommand               Key = "error.unclosed_command"
	EmptyCommand                  Key = "error.empty_command"
	NoTargets                     Key = "list.no_targets"
	ErrReadConfig                 Key = "error.read_config"
	ErrParseConfig                Key = "error.parse_config"
	ErrGenerateConfig             Key = "error.generate_config"
	ErrCreateConfigDirectory      Key = "error.create_config_directory"
	ErrCreateTempConfig           Key = "error.create_temp_config"
	ErrSetConfigPermissions       Key = "error.set_config_permissions"
	ErrWriteTempConfig            Key = "error.write_temp_config"
	ErrSyncTempConfig             Key = "error.sync_temp_config"
	ErrCloseTempConfig            Key = "error.close_temp_config"
	ErrSaveConfig                 Key = "error.save_config"
	ErrCheckConfig                Key = "error.check_config"
	ErrConfigNotFound             Key = "error.config_not_found"
	ErrDetermineConfigDirectory   Key = "error.determine_config_directory"
	ErrResolveConfigPath          Key = "error.resolve_config_path"
	ErrResolveLocalDirectory      Key = "error.resolve_local_directory"
	ErrArchiveConfig              Key = "error.archive_config"
	ErrCheckLocalDirectory        Key = "error.check_local_directory"
	ValidationVersion             Key = "validation.version"
	ValidationProjectRequired     Key = "validation.project_required"
	ValidationProjectID           Key = "validation.project_id"
	ValidationFieldRequired       Key = "validation.field_required"
	ValidationEnvironmentRequired Key = "validation.environment_required"
	ValidationEnvironmentID       Key = "validation.environment_id"
	ValidationAbsolutePath        Key = "validation.absolute_path"
	ValidationBuildRequired       Key = "validation.build_required"
	ValidationArtifactPath        Key = "validation.artifact_path"
	ValidationSSHTarget           Key = "validation.ssh_target"
	ValidationRemotePath          Key = "validation.remote_path"
	ValidationDirtyPolicy         Key = "validation.dirty_policy"
)

type Translator struct {
	language Language
}

func New(language Language) Translator {
	if language != SimplifiedChinese {
		language = English
	}
	return Translator{language: language}
}

func (t Translator) Language() Language { return t.language }

func (t Translator) T(key Key, args ...any) string {
	template, ok := catalogs[t.language][key]
	if !ok {
		template = catalogs[English][key]
	}
	if template == "" {
		template = string(key)
	}
	return fmt.Sprintf(template, args...)
}

type Localizable interface {
	Localize(Translator) string
}

type Error struct {
	Key   Key
	Args  []any
	Cause error
}

func NewError(key Key, args ...any) error {
	return &Error{Key: key, Args: args}
}

func Wrap(key Key, cause error, args ...any) error {
	return &Error{Key: key, Args: args, Cause: cause}
}

func (e *Error) Error() string { return e.Localize(New(English)) }
func (e *Error) Unwrap() error { return e.Cause }
func (e *Error) Localize(translator Translator) string {
	args := append([]any(nil), e.Args...)
	if e.Cause != nil {
		args = append(args, translator.Error(e.Cause))
	}
	return translator.T(e.Key, args...)
}

func (t Translator) Error(err error) string {
	if err == nil {
		return ""
	}
	var localizable Localizable
	if errors.As(err, &localizable) {
		return localizable.Localize(t)
	}
	return err.Error()
}

func Resolve(requested string) (Translator, error) {
	if requested != "" && !strings.EqualFold(requested, "auto") {
		language, ok := normalize(requested)
		if !ok {
			return New(English), fmt.Errorf("unsupported language %q; supported values: auto, en, zh-CN", requested)
		}
		return New(language), nil
	}
	if configured := os.Getenv("DISTSHIP_LANG"); configured != "" && !strings.EqualFold(configured, "auto") {
		language, ok := normalize(configured)
		if !ok {
			return New(English), fmt.Errorf("unsupported DISTSHIP_LANG value %q; supported values: auto, en, zh-CN", configured)
		}
		return New(language), nil
	}
	for _, name := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if language, ok := normalize(os.Getenv(name)); ok {
			return New(language), nil
		}
	}
	return New(English), nil
}

func normalize(value string) (Language, bool) {
	value = strings.TrimSpace(strings.ToLower(value))
	if index := strings.IndexByte(value, '.'); index >= 0 {
		value = value[:index]
	}
	if index := strings.IndexByte(value, '@'); index >= 0 {
		value = value[:index]
	}
	value = strings.ReplaceAll(value, "_", "-")
	switch {
	case value == "en" || strings.HasPrefix(value, "en-"):
		return English, true
	case value == "zh" || value == "zh-cn" || value == "zh-hans" || strings.HasPrefix(value, "zh-hans-"):
		return SimplifiedChinese, true
	default:
		return "", false
	}
}
