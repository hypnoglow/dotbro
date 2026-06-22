package main

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInferAddRepositoryPath_MappingFirstLongestPrefix(t *testing.T) {
	profile := &Profile{data: ProfileData{Mapping: map[string]Destinations{
		"pi/settings.json":           {".pi/agent/settings.json"},
		"pi/deep/existing.json":      {".pi/agent/deep/existing.json"},
		"other/shall-not-match.json": {".other/shall-not-match.json"},
		"pi/deep-even/no-match.json": {".pi/agent/deep-even/no-match.json"},
	}}}

	inf, err := inferAddRepositoryPath(profile, ".pi/agent/deep/new.json")

	require.NoError(t, err)
	assert.Equal(t, "pi", inf.App)
	assert.Equal(t, "pi/deep/new.json", inf.SourceRel)
}

func TestInferAddRepositoryPath_Fallbacks(t *testing.T) {
	tests := []struct {
		name       string
		destRel    string
		wantApp    string
		wantSource string
	}{
		{
			name:       "config app",
			destRel:    ".config/foo/config.toml",
			wantApp:    "foo",
			wantSource: "foo/config.toml",
		},
		{
			name:       "dot app dir",
			destRel:    ".foo/bar/baz.toml",
			wantApp:    "foo",
			wantSource: "foo/bar/baz.toml",
		},
		{
			name:       "application support name",
			destRel:    "Library/Application Support/Ghostty/config",
			wantApp:    "Ghostty",
			wantSource: "Ghostty/config",
		},
		{
			name:       "application support bundle id",
			destRel:    "Library/Application Support/com.mitchellh.ghostty/config",
			wantApp:    "ghostty",
			wantSource: "ghostty/config",
		},
		{
			name:       "single hidden file",
			destRel:    ".foo",
			wantApp:    "foo",
			wantSource: "foo/foo",
		},
		{
			name:       "single rc file",
			destRel:    ".foorc",
			wantApp:    "foo",
			wantSource: "foo/foorc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inf, err := inferAddRepositoryPath(&Profile{}, tt.destRel)
			require.NoError(t, err)
			assert.Equal(t, tt.wantApp, inf.App)
			assert.Equal(t, tt.wantSource, inf.SourceRel)
		})
	}
}

func TestProfileAwareAppDetection(t *testing.T) {
	t.Run("repo structure", func(t *testing.T) {
		dir := t.TempDir()
		require.NoError(t, os.MkdirAll(filepath.Join(dir, "claude", "@profiles", "acrux"), 0700))
		profile := &Profile{data: ProfileData{Directories: Directories{Dotfiles: dir}}}

		assert.True(t, isProfileAwareApp(profile, "claude"))
	})

	t.Run("current mapping", func(t *testing.T) {
		profile := &Profile{data: ProfileData{Mapping: map[string]Destinations{
			"claude/@profiles/acrux/settings.json": {".claude/settings.json"},
		}}}

		assert.True(t, isProfileAwareApp(profile, "claude"))
	})
}

func TestMakeProfileSpecificRepoPath(t *testing.T) {
	assert.Equal(t, "claude/@profiles/acrux/hooks/new.sh", makeProfileSpecificRepoPath("claude/hooks/new.sh", "acrux"))
	assert.Equal(t, "claude/@profiles/acrux/hooks/new.sh", makeProfileSpecificRepoPath("claude/@profiles/antares/hooks/new.sh", "acrux"))
}

func TestApproveAddPlan_ConfigScope(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "dotbro", "@profiles", "acrux", "dotbro.toml")
	other := filepath.Join(dir, "dotbro", "@profiles", "antares", "dotbro.toml")
	require.NoError(t, os.MkdirAll(filepath.Dir(current), 0700))
	require.NoError(t, os.MkdirAll(filepath.Dir(other), 0700))
	require.NoError(t, os.WriteFile(current, []byte("[mapping]\n"), 0600))
	require.NoError(t, os.WriteFile(other, []byte("[mapping]\n"), 0600))

	profile := &Profile{filepath: current, data: ProfileData{Directories: Directories{Dotfiles: dir, Backup: filepath.Join(dir, "backup")}}}
	plan := addPlan{RepoRel: "foo/config.toml", DestinationRel: ".config/foo/config.toml", ConfigScope: addScopeCurrent, ConfigPaths: []string{current}}

	approved, err := approveAddPlan(bytes.NewBufferString("\na\ny\n"), &bytes.Buffer{}, plan, profile, "acrux")

	require.NoError(t, err)
	assert.Equal(t, addScopeAll, approved.ConfigScope)
	assert.Equal(t, []string{current, other}, approved.ConfigPaths)
}

func TestApproveAddPlan_ProfileSpecificForcesCurrentScope(t *testing.T) {
	dir := t.TempDir()
	current := filepath.Join(dir, "dotbro", "@profiles", "acrux", "dotbro.toml")
	profile := &Profile{filepath: current, data: ProfileData{Directories: Directories{Dotfiles: dir, Backup: filepath.Join(dir, "backup")}}}
	plan := addPlan{RepoRel: "foo/config.toml", DestinationRel: ".config/foo/config.toml", ConfigScope: addScopeCurrent, ConfigPaths: []string{current}}

	approved, err := approveAddPlan(bytes.NewBufferString("foo/@profiles/acrux/config.toml\ny\n"), &bytes.Buffer{}, plan, profile, "acrux")

	require.NoError(t, err)
	assert.Equal(t, addPlacementProfileSpecific, approved.Placement)
	assert.Equal(t, addScopeCurrent, approved.ConfigScope)
	assert.Equal(t, []string{current}, approved.ConfigPaths)
}

func TestInsertMappingEntry_AfterSameApp(t *testing.T) {
	content := `[directories]
dotfiles = "/tmp/dotfiles"

[mapping]
"foo/a.toml" = ".config/foo/a.toml"
"bar/config.toml" = ".config/bar/config.toml"
"foo/b.toml" = ".config/foo/b.toml"

[files]
excludes = []
`

	got, err := insertMappingEntry(content, "foo/c.toml", ".config/foo/c.toml")

	require.NoError(t, err)
	assert.Contains(t, got, "\"foo/b.toml\" = \".config/foo/b.toml\"\n\"foo/c.toml\" = \".config/foo/c.toml\"\n\n[files]")
}

func TestInsertMappingEntry_EndOfMappingSectionWhenAppAbsent(t *testing.T) {
	content := `[mapping]
"bar/config.toml" = ".config/bar/config.toml"

[files]
excludes = []
`

	got, err := insertMappingEntry(content, "foo/config.toml", ".config/foo/config.toml")

	require.NoError(t, err)
	assert.Contains(t, got, "\"bar/config.toml\" = \".config/bar/config.toml\"\n\n\"foo/config.toml\" = \".config/foo/config.toml\"\n[files]")
}

func TestInsertMappingEntry_MissingMappingSection(t *testing.T) {
	_, err := insertMappingEntry("[directories]\n", "foo/config.toml", ".config/foo/config.toml")

	assert.Error(t, err)
}

func TestValidateTOMLProfileForAdd(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "dotbro.json")
	tomlWithoutMapping := filepath.Join(dir, "dotbro.toml")
	tomlWithMapping := filepath.Join(dir, "with-mapping.toml")
	require.NoError(t, os.WriteFile(jsonPath, []byte(`{"mapping":{}}`), 0600))
	require.NoError(t, os.WriteFile(tomlWithoutMapping, []byte("[directories]\n"), 0600))
	require.NoError(t, os.WriteFile(tomlWithMapping, []byte("[mapping]\n"), 0600))

	assert.Error(t, validateTOMLProfileForAdd(jsonPath))
	assert.Error(t, validateTOMLProfileForAdd(tomlWithoutMapping))
	assert.NoError(t, validateTOMLProfileForAdd(tomlWithMapping))
}

func TestDestinationRelativePathRejectsOutsideDestination(t *testing.T) {
	dir := t.TempDir()
	outside := filepath.Join(t.TempDir(), "file")

	_, err := destinationRelativePath(dir, outside)

	assert.Error(t, err)
}

func TestIsInteractiveTerminalRejectsNonTTY(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "input")
	require.NoError(t, err)
	defer file.Close()

	assert.False(t, isInteractiveTerminal(file))
}

func TestBackupPathUsesDestinationRelativePath(t *testing.T) {
	backupRoot := filepath.Join(t.TempDir(), "backup")
	destRel := ".config/foo/config.toml"

	backupPath := filepath.Join(backupRoot, filepath.FromSlash(destRel))

	assert.Equal(t, filepath.Join(backupRoot, ".config", "foo", "config.toml"), backupPath)
}
