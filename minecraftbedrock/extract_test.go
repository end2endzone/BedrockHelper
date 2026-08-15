package minecraftbedrock

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindManifestsRelativePathInAddon(t *testing.T) {
	t.Run("addon with two packs", func(t *testing.T) {
		actualNames, err := FindManifestsRelativePathInAddon(getAddonFixturePath(t, "foobar.mcaddon"))
		require.NoError(t, err)

		expectedNames := []string{
			"foobar_BP/manifest.json",
			"foobar_RP/manifest.json",
		}
		require.ElementsMatch(t, expectedNames, actualNames)
	})

	t.Run("standalone mcpack", func(t *testing.T) {
		actualNames, err := FindManifestsRelativePathInAddon(getAddonFixturePath(t, "solo.mcpack"))
		require.NoError(t, err)

		expectedNames := []string{"manifest.json"}
		require.Equal(t, expectedNames, actualNames)
	})

	t.Run("zip with no manifest", func(t *testing.T) {
		_, err := FindManifestsRelativePathInAddon(getAddonFixturePath(t, "zip_with_no_manifest.zip"))
		require.Error(t, err)
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := FindManifestsRelativePathInAddon("/tmp/nope.mcaddon")
		require.Error(t, err)
	})
}

func TestExtractZip(t *testing.T) {
	dest := t.TempDir()
	err := ExtractZip(getAddonFixturePath(t, "foobar.mcaddon"), dest)
	require.NoError(t, err)

	for _, rel := range []string{
		"foobar_BP/manifest.json",
		"foobar_BP/items/coin.json",
		"foobar_RP/manifest.json",
		"foobar_RP/textures/items/coin.png",
	} {
		path := filepath.Join(dest, filepath.FromSlash(rel))
		assertFileExists(t, path)
	}
}

func TestReadZipEntry(t *testing.T) {
	// Read a known manifest.json file
	data, err := readZipEntry(getAddonFixturePath(t, "solo.mcpack"), "manifest.json")
	require.NoError(t, err)

	m, err := LoadManifestFromBytes(data)
	require.NoError(t, err, "failed to parse extracted manifest")
	require.Equal(t, "Solo RP", m.Header.Name)

	// Test invalid path
	_, err = readZipEntry(getAddonFixturePath(t, "solo.mcpack"), "does/not/exist.json")
	require.Error(t, err)
}
