package libminecraftbedrock

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
	require.NoError(t, err, "failed to get working directory")

	// This file lives in <module>/libminecraftbedrock, so testdata is a sibling of that directory.
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
// This is required to prevent calling functions that would modify the directory. It makes sure we never affect the testdata files under CM.
// Returns the path to the new temporary directory. t.TempDir() takes care of cleanup automatically.
func copyServerFixture(t *testing.T, name string) string {
	t.Helper()
	src := getServerFixturePath(t, name)
	dst := filepath.Join(t.TempDir(), name)
	err := copyDir(src, dst)
	require.NoError(t, err, "failed to copy server fixture %q", name)

	return dst
}

// copyAddonFixture copies the given testdata addon file into a temporary directory.
// This is required to prevent calling functions that would modify the addon. It makes sure we never affect the testdata files under CM.
// Returns the path to the new temporary file. t.TempDir() takes care of cleanup automatically.
func copyAddonFixture(t *testing.T, name string) string {
	t.Helper()
	src := getAddonFixturePath(t, name)
	dst := filepath.Join(t.TempDir(), filepath.Base(src)) // put the file directly in TempDir.
	err := copyFile(src, dst)
	require.NoError(t, err, "failed to copy addon fixture %q", name)
	return dst
}

// getTempServer copies the given testdata server fixture into a temporary directory and returns it as a valid Server.
func getTempServer(t *testing.T, name string) *Server {
	t.Helper()
	dir := copyServerFixture(t, name)
	server, err := GetServer(dir)
	require.NoError(t, err, "failed to get server for fixture %q", name)
	return server
}

// getTempActiveWorld copies the given testdata server fixture into a temporary directory and returns its active World.
func getTempActiveWorld(t *testing.T, name string) *World {
	t.Helper()
	server := getTempServer(t, name)
	world, err := server.ActiveWorld()
	require.NoError(t, err, "failed to get active world for fixture %q", name)
	return world
}

func TestCopyAddonFixture(t *testing.T) {
	tempAddonPath := copyAddonFixture(t, "foobar.mcaddon")
	require.NotEmpty(t, tempAddonPath, "expected a non-empty temporary addon path")

	_, err := os.Stat(tempAddonPath)
	require.NoError(t, err, "expected the copied addon file to exist")
}

func TestCopyServerFixture(t *testing.T) {
	tempServerPath := copyServerFixture(t, "server")
	require.NotEmpty(t, tempServerPath, "expected a non-empty temporary server path")

	info, err := os.Stat(tempServerPath)
	require.NoError(t, err, "expected the copied server directory to exist")
	require.True(t, info.IsDir())
}
