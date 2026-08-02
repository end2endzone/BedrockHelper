package libbedrockpacks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// testdataDir returns the absolute path to the testdata directory located at the project's root directory.
// It returns the directory location regardless of which package's test binary is running.
func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)

	// This file lives in <module>/libbedrockpacks, so testdata is a sibling
	// of that directory.
	return filepath.Join(wd, "..", "testdata")
}

func getAddonFixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testdataDir(t), "addons", name)
}

func getServerFixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testdataDir(t), "servers", name)
}

// copyServerFixture copies the given testdata server directory into a temporary directory.
// This is required to prevent calling function that would modify the directory. It makes sure we never affect the testdata files under CM.
// Returns the path to the new temporary directory. The caller is reponsible to delete the returned temporary directory.
func copyServerFixture(t *testing.T, name string) string {
	t.Helper()
	src := getServerFixturePath(t, name)
	dst := filepath.Join(t.TempDir(), name)
	err := copyDir(src, dst)
	require.NoError(t, err, "failed to copy server fixture %q", name)

	return dst
}

// copyAddonFixture copies the given testdata addon file into a temporary directory.
// This is required to prevent calling function that would modify the addon. It makes sure we never affect the testdata files under CM.
// Returns the path to the new temporary file. The caller is reponsible to delete the returned temporary file.
func copyAddonFixture(t *testing.T, name string) string {
	t.Helper()
	src := getAddonFixturePath(t, name)
	dst := filepath.Join(t.TempDir(), filepath.Base(src)) // put the file directly in TempDir.
	err := copyFile(src, dst)
	require.NoError(t, err, "failed to copy addon fixture %q", name)

	return dst
}

func TestCopyAddonFixture(t *testing.T) {
	tempAddonPath := copyAddonFixture(t, "foobar.mcaddon")
	defer os.Remove(tempAddonPath)

	require.NotEqual(t, "", tempAddonPath)
}
