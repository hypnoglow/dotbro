package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type addPlacement string

const (
	addPlacementCommon          addPlacement = "common"
	addPlacementProfileSpecific addPlacement = "profile-specific"
	profileMarker               string       = "@profiles"
)

type addScope string

const (
	addScopeCurrent addScope = "current"
	addScopeAll     addScope = "all"
)

type addPlan struct {
	InputPath      string
	DestinationRel string
	App            string
	RepoRel        string
	Placement      addPlacement
	BackupPath     string
	SymlinkDest    string
	MappingSource  string
	MappingDest    string
	ConfigScope    addScope
	ConfigPaths    []string
}

type addInference struct {
	App       string
	SourceRel string
}

func (app *App) addAction(ctx context.Context, filename string) error {
	if !isInteractiveTerminal(os.Stdin) {
		return errors.New("dotbro add requires an interactive terminal")
	}

	inputAbs, err := filepath.Abs(filename)
	if err != nil {
		return err
	}
	inputAbs = filepath.Clean(inputAbs)

	fileInfo, err := os.Lstat(inputAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%s: no such file or directory", filename)
		}
		return err
	}

	if fileInfo.Mode()&os.ModeSymlink == os.ModeSymlink {
		return fmt.Errorf("cannot add file %s: symlinks are not supported", filename)
	}

	if fileInfo.Mode().IsDir() {
		return fmt.Errorf("cannot add dir %s: directories are not supported yet", filename)
	}

	if !fileInfo.Mode().IsRegular() {
		return fmt.Errorf("cannot add file %s: only regular files are supported", filename)
	}

	if err = validateTOMLProfileForAdd(app.profile.Filepath()); err != nil {
		return err
	}

	destRel, err := destinationRelativePath(app.profile.DestinationDir(), inputAbs)
	if err != nil {
		return err
	}

	profileName, hasProfileName := profileNameFromConfigPath(app.profile.Filepath())
	inf, err := inferAddRepositoryPath(app.profile, destRel)
	if err != nil {
		return err
	}

	repoRel := inf.SourceRel
	placement := addPlacementCommon
	if isProfileAwareApp(app.profile, inf.App) {
		if !hasProfileName {
			return fmt.Errorf("selected config %s is not under @profiles/<name>/; cannot choose profile-specific placement automatically", app.profile.Filepath())
		}
		repoRel = makeProfileSpecificRepoPath(inf.SourceRel, profileName)
		placement = addPlacementProfileSpecific
	}

	plan := addPlan{
		InputPath:      inputAbs,
		DestinationRel: destRel,
		App:            inf.App,
		RepoRel:        repoRel,
		Placement:      placement,
		BackupPath:     filepath.Join(app.profile.BackupDir(), filepath.FromSlash(destRel)),
		SymlinkDest:    inputAbs,
		MappingSource:  repoRel,
		MappingDest:    destRel,
		ConfigScope:    addScopeCurrent,
		ConfigPaths:    []string{app.profile.Filepath()},
	}

	approvedPlan, err := approveAddPlan(os.Stdin, os.Stdout, plan, app.profile, profileName)
	if err != nil {
		return err
	}
	plan = approvedPlan

	if err = validateAddPlanBeforeChanges(plan, app.profile); err != nil {
		return err
	}

	app.logger.DebugContext(ctx, "Adding file to dotfiles repository",
		slog.String("src", plan.InputPath),
		slog.String("repo", plan.RepoRel),
		slog.String("dst", plan.DestinationRel))

	if err = Copy(osfs, plan.InputPath, plan.BackupPath); err != nil {
		return fmt.Errorf("cannot backup file %s: %s", plan.InputPath, err)
	}
	app.logger.InfoContext(ctx, "backup",
		slog.String("status", "→"),
		slog.String("src", plan.InputPath),
		slog.String("dst", plan.BackupPath))

	repoAbs := filepath.Join(app.profile.DotfilesDir(), filepath.FromSlash(plan.RepoRel))
	if err = os.MkdirAll(filepath.Dir(repoAbs), 0700); err != nil {
		return err
	}
	if err = os.Rename(plan.InputPath, repoAbs); err != nil {
		return err
	}

	linker := NewLinker(osfs, app.logger)
	if err = linker.SetSymlink(repoAbs, plan.SymlinkDest); err != nil {
		return err
	}

	for _, configPath := range plan.ConfigPaths {
		if err = insertMappingEntryIntoFile(configPath, plan.MappingSource, plan.MappingDest); err != nil {
			return fmt.Errorf("edit mapping in %s: %w", configPath, err)
		}
	}

	return nil
}

func (app *App) getCurrentProfilePath(ctx context.Context, profileArg any) string {
	if profileArg != nil {
		profilePath, err := filepath.Abs(profileArg.(string))
		if err != nil {
			app.logger.ErrorContext(ctx, "Bad profile path", slog.Any("error", err))
			app.exit(1)
		}

		cfg := NewConfig(app.logger, defaultConfigFilepath, defaultLegacyConfigFilepath)
		if err := cfg.Load(ctx); err != nil {
			app.logger.ErrorContext(ctx, "Error reading config", slog.Any("error", err))
			app.exit(1)
		}
		cfg.AddProfile(profilePath)
		if err := cfg.Save(ctx); err != nil {
			app.logger.ErrorContext(ctx, "Cannot save config", slog.Any("error", err))
			app.exit(1)
		}
		return profilePath
	}

	cfg := NewConfig(app.logger, defaultConfigFilepath, defaultLegacyConfigFilepath)
	if err := cfg.Load(ctx); err != nil {
		app.logger.ErrorContext(ctx, "Error reading config", slog.Any("error", err))
		app.exit(1)
	}

	host, err := currentLocalHostName()
	if err != nil {
		app.logger.ErrorContext(ctx, "Cannot infer current profile from hostname", slog.Any("error", err))
		app.logger.InfoContext(ctx, "Pass --config to select a dotbro profile explicitly.", slog.String("tip", "TIP"))
		app.exit(1)
	}

	if profilePath, ok := findCurrentProfileForHost(cfg.GetProfilePaths(), host); ok {
		app.logger.DebugContext(ctx, "Using current profile inferred from hostname", slog.String("host", host), slog.String("path", profilePath))
		return profilePath
	}

	app.logger.ErrorContext(ctx, "Cannot infer current profile from hostname", slog.String("host", host))
	app.logger.InfoContext(ctx, "Pass --config to select a dotbro profile explicitly.", slog.String("tip", "TIP"))
	app.exit(1)
	return ""
}

func (app *App) loadProfile(ctx context.Context, profilePath string) error {
	app.logger.DebugContext(ctx, "Loading profile", slog.String("path", profilePath))
	profile, err := NewProfile(profilePath)
	if err != nil {
		return err
	}
	app.profile = profile

	if err = os.MkdirAll(app.profile.BackupDir(), 0700); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error creating backup directory: %w", err)
	}

	app.logger.DebugContext(ctx, "Profile directories",
		slog.String("dotfiles", app.profile.DotfilesDir()),
		slog.String("sources", app.profile.SourcesDir()),
		slog.String("destination", app.profile.DestinationDir()),
		slog.String("backup", app.profile.BackupDir()))

	return nil
}

func isInteractiveTerminal(file *os.File) bool {
	info, err := file.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func validateTOMLProfileForAdd(profilePath string) error {
	if filepath.Ext(profilePath) != ".toml" {
		return fmt.Errorf("dotbro add can edit only TOML profiles automatically: %s", profilePath)
	}
	content, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}
	if !hasMappingSection(string(content)) {
		return fmt.Errorf("profile %s does not contain an existing [mapping] section", profilePath)
	}
	return nil
}

func destinationRelativePath(destinationDir, inputPath string) (string, error) {
	destinationAbs, err := filepath.Abs(destinationDir)
	if err != nil {
		return "", err
	}
	destinationAbs = filepath.Clean(destinationAbs)

	rel, err := filepath.Rel(destinationAbs, inputPath)
	if err != nil {
		return "", err
	}
	if rel == "." || rel == "" || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", fmt.Errorf("input file %s is outside directories.destination %s", inputPath, destinationAbs)
	}
	return filepath.ToSlash(rel), nil
}

func inferAddRepositoryPath(profile *Profile, destRel string) (addInference, error) {
	if inf, ok := inferFromCurrentProfileMapping(profile.Data().Mapping, destRel); ok {
		return inf, nil
	}
	if inf, ok := inferFromPathFallback(destRel); ok {
		return inf, nil
	}
	return addInference{}, fmt.Errorf("cannot infer repository path for destination %s", destRel)
}

func inferFromCurrentProfileMapping(mapping map[string]Destinations, destRel string) (addInference, bool) {
	type candidate struct {
		destPrefix string
		sourceRel  string
	}

	var best candidate
	bestLen := -1
	for src, dsts := range mapping {
		for _, dst := range dsts {
			destPrefix := pathDirSlash(dst)
			if destPrefix == "." || destPrefix == "" {
				continue
			}
			if !pathHasPrefix(destRel, destPrefix) {
				continue
			}
			prefixLen := len(splitSlash(destPrefix))
			if prefixLen <= bestLen {
				continue
			}
			rest := slashRel(destPrefix, destRel)
			sourcePrefix := pathDirSlash(src)
			if sourcePrefix == "." {
				sourcePrefix = ""
			}
			sourceRel := joinSlash(sourcePrefix, rest)
			if sourceRel == "" {
				sourceRel = src
			}
			best = candidate{destPrefix: destPrefix, sourceRel: sourceRel}
			bestLen = prefixLen
		}
	}

	if bestLen == -1 {
		return addInference{}, false
	}
	app := appNameFromRepoPath(best.sourceRel)
	if app == "" {
		return addInference{}, false
	}
	_ = best.destPrefix
	return addInference{App: app, SourceRel: cleanSlash(best.sourceRel)}, true
}

func inferFromPathFallback(destRel string) (addInference, bool) {
	parts := splitSlash(destRel)
	if len(parts) >= 3 && parts[0] == ".config" {
		app := parts[1]
		return addInference{App: app, SourceRel: joinSlash(app, joinSlash(parts[2:]...))}, true
	}
	if len(parts) >= 3 && parts[0] == "Library" && parts[1] == "Application Support" {
		app := normalizeApplicationSupportName(parts[2])
		return addInference{App: app, SourceRel: joinSlash(app, joinSlash(parts[3:]...))}, true
	}
	if len(parts) >= 2 && strings.HasPrefix(parts[0], ".") && parts[0] != "." && parts[0] != ".." {
		app := strings.TrimPrefix(parts[0], ".")
		return addInference{App: app, SourceRel: joinSlash(app, joinSlash(parts[1:]...))}, true
	}
	if len(parts) == 1 && strings.HasPrefix(parts[0], ".") && len(parts[0]) > 1 {
		name := strings.TrimPrefix(parts[0], ".")
		app := name
		if strings.HasSuffix(name, "rc") && len(name) > 2 {
			app = strings.TrimSuffix(name, "rc")
		}
		return addInference{App: app, SourceRel: joinSlash(app, name)}, true
	}
	return addInference{}, false
}

func normalizeApplicationSupportName(name string) string {
	name = strings.TrimPrefix(name, ".")
	if strings.Contains(name, ".") {
		parts := strings.Split(name, ".")
		return parts[len(parts)-1]
	}
	return name
}

func isProfileAwareApp(profile *Profile, app string) bool {
	if app == "" {
		return false
	}
	if _, err := os.Stat(filepath.Join(profile.DotfilesDir(), app, profileMarker)); err == nil {
		return true
	}
	for src := range profile.Data().Mapping {
		parts := splitSlash(src)
		if len(parts) >= 3 && parts[0] == app && parts[1] == profileMarker {
			return true
		}
	}
	return false
}

func makeProfileSpecificRepoPath(repoRel, profileName string) string {
	parts := splitSlash(repoRel)
	if len(parts) == 0 {
		return repoRel
	}
	app := parts[0]
	rest := parts[1:]
	if len(rest) >= 2 && rest[0] == profileMarker {
		rest = rest[2:]
	}
	out := append([]string{app, profileMarker, profileName}, rest...)
	return joinSlash(out...)
}

func approveAddPlan(in io.Reader, out io.Writer, plan addPlan, profile *Profile, currentProfileName string) (addPlan, error) {
	printAddPlan(out, plan)

	reader := bufio.NewReader(in)
	fmt.Fprintf(out, "Final repo path [%s]: ", plan.RepoRel)
	line, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return addPlan{}, err
	}
	line = strings.TrimSpace(line)
	if line != "" {
		cleaned, err := validateUserRepoPath(line)
		if err != nil {
			return addPlan{}, err
		}
		plan.RepoRel = cleaned
		plan.MappingSource = cleaned
		plan.App = appNameFromRepoPath(cleaned)
	}

	if repoPathContainsProfile(plan.RepoRel, currentProfileName) {
		plan.Placement = addPlacementProfileSpecific
		plan.ConfigScope = addScopeCurrent
		plan.ConfigPaths = []string{profile.Filepath()}
	} else {
		plan.Placement = addPlacementCommon
		fmt.Fprint(out, "Update config scope? [c]urrent/[a]ll profiles (default current): ")
		scopeLine, err := reader.ReadString('\n')
		if err != nil && !errors.Is(err, io.EOF) {
			return addPlan{}, err
		}
		scopeLine = strings.ToLower(strings.TrimSpace(scopeLine))
		switch scopeLine {
		case "", "c", "current":
			plan.ConfigScope = addScopeCurrent
			plan.ConfigPaths = []string{profile.Filepath()}
		case "a", "all":
			paths, err := allProfileConfigPaths(profile)
			if err != nil {
				return addPlan{}, err
			}
			plan.ConfigScope = addScopeAll
			plan.ConfigPaths = paths
		default:
			return addPlan{}, fmt.Errorf("unknown config scope %q", scopeLine)
		}
	}

	plan.BackupPath = filepath.Join(profile.BackupDir(), filepath.FromSlash(plan.DestinationRel))
	plan.MappingDest = plan.DestinationRel

	fmt.Fprintln(out)
	fmt.Fprintln(out, "Final approved plan:")
	printAddPlan(out, plan)
	fmt.Fprint(out, "Proceed? [y/N]: ")
	confirm, err := reader.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return addPlan{}, err
	}
	confirm = strings.ToLower(strings.TrimSpace(confirm))
	if confirm != "y" && confirm != "yes" {
		return addPlan{}, errors.New("aborted")
	}

	return plan, nil
}

func printAddPlan(out io.Writer, plan addPlan) {
	fmt.Fprintln(out, "dotbro add plan:")
	fmt.Fprintf(out, "  input file:           %s\n", plan.InputPath)
	fmt.Fprintf(out, "  destination-relative: %s\n", plan.DestinationRel)
	fmt.Fprintf(out, "  inferred app:         %s\n", plan.App)
	fmt.Fprintf(out, "  final repo path:      %s\n", plan.RepoRel)
	fmt.Fprintf(out, "  placement:            %s\n", plan.Placement)
	fmt.Fprintf(out, "  backup path:          %s\n", plan.BackupPath)
	fmt.Fprintf(out, "  symlink destination:  %s\n", plan.SymlinkDest)
	fmt.Fprintf(out, "  mapping entry:        %q = %q\n", plan.MappingSource, plan.MappingDest)
	fmt.Fprintln(out, "  target config file(s):")
	for _, p := range plan.ConfigPaths {
		fmt.Fprintf(out, "    - %s\n", p)
	}
}

func validateUserRepoPath(value string) (string, error) {
	value = filepath.ToSlash(filepath.Clean(value))
	if value == "." || value == "" || strings.HasPrefix(value, "../") || value == ".." || strings.HasPrefix(value, "/") {
		return "", fmt.Errorf("repo path must be relative: %s", value)
	}
	return value, nil
}

func repoPathContainsProfile(repoRel, currentProfileName string) bool {
	parts := splitSlash(repoRel)
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == profileMarker {
			if currentProfileName == "" {
				return true
			}
			return parts[i+1] == currentProfileName
		}
	}
	return false
}

func allProfileConfigPaths(profile *Profile) ([]string, error) {
	pattern := filepath.Join(profile.DotfilesDir(), "dotbro", profileMarker, "*", "dotbro.toml")
	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	sort.Strings(paths)
	if len(paths) == 0 {
		return nil, fmt.Errorf("no profile configs found under %s", pattern)
	}
	return paths, nil
}

func validateAddPlanBeforeChanges(plan addPlan, profile *Profile) error {
	repoAbs := filepath.Join(profile.DotfilesDir(), filepath.FromSlash(plan.RepoRel))
	if _, err := os.Lstat(repoAbs); err == nil {
		return fmt.Errorf("repo path already exists: %s", repoAbs)
	} else if !os.IsNotExist(err) {
		return err
	}

	for _, configPath := range plan.ConfigPaths {
		if err := validateTOMLProfileForAdd(configPath); err != nil {
			return err
		}
	}
	return nil
}

func insertMappingEntryIntoFile(profilePath, source, dest string) error {
	content, err := os.ReadFile(profilePath)
	if err != nil {
		return err
	}
	updated, err := insertMappingEntry(string(content), source, dest)
	if err != nil {
		return err
	}
	return os.WriteFile(profilePath, []byte(updated), 0600)
}

var tomlTableRegexp = regexp.MustCompile(`^\s*\[[^\[][^\]]*\]\s*(?:#.*)?$`)
var tomlMappingEntryRegexp = regexp.MustCompile(`^\s*("(?:\\.|[^"])*")\s*=`)

func hasMappingSection(content string) bool {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(line))
		if trimmed == "[mapping]" {
			return true
		}
	}
	return false
}

func insertMappingEntry(content, source, dest string) (string, error) {
	lines := strings.SplitAfter(content, "\n")
	sectionStart := -1
	sectionEnd := len(lines)
	for i, line := range lines {
		trimmed := strings.TrimSpace(stripInlineComment(strings.TrimRight(line, "\r\n")))
		if trimmed == "[mapping]" {
			sectionStart = i
			break
		}
	}
	if sectionStart == -1 {
		return "", errors.New("missing [mapping] section")
	}
	for i := sectionStart + 1; i < len(lines); i++ {
		lineNoNewline := strings.TrimRight(lines[i], "\r\n")
		if tomlTableRegexp.MatchString(strings.TrimSpace(lineNoNewline)) {
			sectionEnd = i
			break
		}
	}

	app := appNameFromRepoPath(source)
	insertIndex := sectionEnd
	for i := sectionStart + 1; i < sectionEnd; i++ {
		key, ok := mappingEntryKey(lines[i])
		if !ok {
			continue
		}
		if appNameFromRepoPath(key) == app {
			insertIndex = i + 1
		}
	}

	entry := fmt.Sprintf("%s = %s\n", quoteTOMLString(source), quoteTOMLString(dest))
	before := strings.Join(lines[:insertIndex], "")
	after := strings.Join(lines[insertIndex:], "")
	if before != "" && !strings.HasSuffix(before, "\n") {
		before += "\n"
	}
	return before + entry + after, nil
}

func mappingEntryKey(line string) (string, bool) {
	matches := tomlMappingEntryRegexp.FindStringSubmatch(line)
	if len(matches) != 2 {
		return "", false
	}
	key, err := strconv.Unquote(matches[1])
	if err != nil {
		return "", false
	}
	return key, true
}

func quoteTOMLString(value string) string {
	return strconv.Quote(value)
}

func stripInlineComment(line string) string {
	inString := false
	escaped := false
	for i, r := range line {
		if escaped {
			escaped = false
			continue
		}
		if r == '\\' && inString {
			escaped = true
			continue
		}
		if r == '"' {
			inString = !inString
			continue
		}
		if r == '#' && !inString {
			return line[:i]
		}
	}
	return line
}

func profileNameFromConfigPath(configPath string) (string, bool) {
	parts := splitSlash(filepath.ToSlash(filepath.Clean(configPath)))
	for i := 0; i+1 < len(parts); i++ {
		if parts[i] == profileMarker && parts[i+1] != "" {
			return parts[i+1], true
		}
	}
	return "", false
}

func currentLocalHostName() (string, error) {
	out, err := exec.CommandContext(context.Background(), "scutil", "--get", "LocalHostName").Output()
	if err != nil {
		return "", fmt.Errorf("run scutil --get LocalHostName: %w", err)
	}
	host := strings.TrimSpace(string(out))
	if host == "" {
		return "", errors.New("scutil --get LocalHostName returned empty hostname")
	}
	return host, nil
}

func findCurrentProfileForHost(profilePaths []string, host string) (string, bool) {
	needle := filepath.ToSlash(filepath.Join(profileMarker, host, "dotbro.toml"))
	for _, p := range profilePaths {
		if strings.Contains(filepath.ToSlash(filepath.Clean(p)), needle) {
			return p, true
		}
	}
	return "", false
}

func cleanSlash(p string) string {
	return filepath.ToSlash(filepath.Clean(filepath.FromSlash(p)))
}

func splitSlash(p string) []string {
	p = cleanSlash(p)
	if p == "." || p == "" {
		return nil
	}
	return strings.Split(p, "/")
}

func joinSlash(parts ...string) string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, splitSlash(part)...)
	}
	return strings.Join(out, "/")
}

func pathDirSlash(p string) string {
	return filepath.ToSlash(filepath.Dir(filepath.FromSlash(p)))
}

func pathHasPrefix(p, prefix string) bool {
	p = cleanSlash(p)
	prefix = cleanSlash(prefix)
	if prefix == "." || prefix == "" {
		return true
	}
	return p == prefix || strings.HasPrefix(p, prefix+"/")
}

func slashRel(base, target string) string {
	baseParts := splitSlash(base)
	targetParts := splitSlash(target)
	if len(targetParts) < len(baseParts) {
		return ""
	}
	return strings.Join(targetParts[len(baseParts):], "/")
}

func appNameFromRepoPath(repoRel string) string {
	parts := splitSlash(repoRel)
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}
