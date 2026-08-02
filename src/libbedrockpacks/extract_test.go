package libbedrockpacks

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindManifestsInAddon(t *testing.T) {
	t.Run("bundle with master + two packs", func(t *testing.T) {
		got, err := FindManifestsRelativePathInAddon(getAddonFixturePath(t, "foobar.mcaddon"))
		require.NoError(t, err)

		sort.Strings(got)
		want := []string{"foobar_BP/manifest.json", "foobar_RP/manifest.json"}

		require.Equal(t, want, got)
	})

	t.Run("standalone mcpack", func(t *testing.T) {
		got, err := FindManifestsRelativePathInAddon(getAddonFixturePath(t, "solo.mcpack"))
		require.NoError(t, err)

		sort.Strings(got)
		want := []string{"manifest.json"}

		require.Equal(t, want, got)
	})

	t.Run("zip with no manifest", func(t *testing.T) {
		_, err := FindManifestsRelativePathInAddon(getAddonFixturePath(t, "no_manifest.zip"))
		require.Error(t, err, "expected error for zip with no manifest.json, got nil")
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := FindManifestsRelativePathInAddon("/tmp/nope.mcaddon")
		require.Error(t, err, "expected error for nonexistent file, got nil")
	})
}

func TestExtractAddon(t *testing.T) {
	dest := t.TempDir()
	err := ExtractZip(getAddonFixturePath(t, "foobar.mcaddon"), dest)
	require.NoError(t, err)

	for _, rel := range []string{
		"foobar_BP/manifest.json",
		"foobar_BP/items/coin.json",
		"foobar_RP/manifest.json",
		"foobar_RP/textures/items/coin.png",
	} {
		_, err := os.Stat(filepath.Join(dest, filepath.FromSlash(rel)))
		require.NoError(t, err)
	}
}

func TestReadZipEntry(t *testing.T) {
	data, err := readZipEntry(getAddonFixturePath(t, "solo.mcpack"), "manifest.json")
	require.NoError(t, err)

	m, err := LoadManifestFromBytes(data)
	require.NoError(t, err)
	require.Equal(t, "Solo RP", m.Header.Name)

	_, err = readZipEntry(getAddonFixturePath(t, "solo.mcpack"), "does/not/exist.json")
	require.Error(t, err)
}
