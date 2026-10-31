package render_test

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/OZON08/openvpn-server-config/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type data struct{ Name string }

func TestTextRendersValues(t *testing.T) {
	txt, err := render.Text("hello {{ .Name }}", data{"vpn"})
	require.NoError(t, err)
	assert.Equal(t, "hello vpn", txt)
}

func TestTextRejectsBrokenTemplate(t *testing.T) {
	_, err := render.Text("{{ ", data{})
	assert.Error(t, err)
}

func TestRequireSingleLineRejectsLineBreaks(t *testing.T) {
	for _, value := range []string{"a\nb", "a\rb"} {
		err := render.RequireSingleLine(data{value})
		require.Error(t, err, "%q", value)
		assert.Contains(t, err.Error(), "Name")
	}
}

func TestRequireSingleLineAllowsNamedFields(t *testing.T) {
	assert.NoError(t, render.RequireSingleLine(data{"a\nb"}, "Name"))
}

func writeTemplate(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "in.tpl")
	require.NoError(t, os.WriteFile(path, []byte("secret {{ .Name }}"), 0600))
	return path
}

func TestSaveToFileCreatesOwnerOnlyFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.conf")
	require.NoError(t, render.SaveToFile(writeTemplate(t), data{"key"}, dest))

	info, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
}

// A container that bind mounts a single file keeps seeing the old inode if the
// file is replaced, so the server would silently ignore every config change.
func TestSaveToFileKeepsInodeOfExistingFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "server.conf")
	require.NoError(t, os.WriteFile(dest, []byte("old"), 0600))
	before, err := os.Stat(dest)
	require.NoError(t, err)

	require.NoError(t, render.SaveToFile(writeTemplate(t), data{"key"}, dest))

	after, err := os.Stat(dest)
	require.NoError(t, err)
	assert.True(t, os.SameFile(before, after), "file must be updated in place")
}

func TestSaveToFileLeavesNoTempFiles(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, render.SaveToFile(writeTemplate(t), data{"key"}, filepath.Join(dir, "out.conf")))

	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	assert.Len(t, entries, 1)
}

func TestSaveToFileTightensExistingFile(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.conf")
	require.NoError(t, os.WriteFile(dest, []byte("a much longer previous content"), 0644))
	require.NoError(t, os.Chmod(dest, 0644))

	require.NoError(t, render.SaveToFile(writeTemplate(t), data{"key"}, dest))

	info, err := os.Stat(dest)
	require.NoError(t, err)
	assert.Equal(t, os.FileMode(0600), info.Mode().Perm())
	content, err := os.ReadFile(dest)
	require.NoError(t, err)
	assert.Equal(t, "secret key", string(content))
}

type invalid struct{}

func (invalid) Validate() error { return errors.New("rejected") }

func TestTextRunsValidator(t *testing.T) {
	_, err := render.Text("static", invalid{})
	assert.EqualError(t, err, "rejected")
}

func TestSaveToFileRunsValidator(t *testing.T) {
	dest := filepath.Join(t.TempDir(), "out.conf")

	assert.EqualError(t, render.SaveToFile(writeTemplate(t), invalid{}, dest), "rejected")
	assert.NoFileExists(t, dest)
}

func TestRequireKeyMaterialRejectsInlineTags(t *testing.T) {
	assert.Error(t, render.RequireKeyMaterial("ok", "a\n</key>"))
	assert.NoError(t, render.RequireKeyMaterial("-----BEGIN CERTIFICATE-----\nabc+/=\n-----END CERTIFICATE-----\n"))
}
