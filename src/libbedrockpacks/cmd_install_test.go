package libbedrockpacks

import (
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallAddon_Bundle(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 2, len(installedPacks))

	var kinds []string
	for _, pack := range installedPacks {
		kinds = append(kinds, pack.KindSafe().String())
		_, err := os.Stat(filepath.Join(pack.Path, "manifest.json"))
		require.NoError(t, err)
	}

	sort.Strings(kinds)
	require.Equal(t, "BehaviorPack", kinds[0])
	require.Equal(t, "ResourcePack", kinds[1])

	// Verify both registry files were updated.
	worldDir, _ := FindActiveWorldDir(tempServerDir)
	bpEntries, err := readRegistry(worldDir, BehaviorPack)
	require.NoError(t, err)
	require.Equal(t, 1, len(bpEntries))

	rpEntries, err := readRegistry(worldDir, ResourcePack)
	require.NoError(t, err)
	require.Equal(t, 1, len(rpEntries))

	// The master link manifest.json must NOT itself be treated as an installed pack.
	for _, p := range installedPacks {
		require.True(t, p.Name() == "Foobar BP" || p.Name() == "Foobar RP")
	}
}

func TestInstallAddon_StandaloneMcpack(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "solo.mcpack"), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(installedPacks))
	require.Equal(t, ResourcePack, installedPacks[0].KindSafe())
	require.Equal(t, "Solo RP", installedPacks[0].Name())
}

func TestInstallAddon_ReinstallReplacesExisting(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)

	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 2, len(installedPacks))

	worldDir, _ := FindActiveWorldDir(tempServerDir)
	bpEntries, _ := readRegistry(worldDir, BehaviorPack)
	require.Equal(t, 1, len(bpEntries))
}

func TestInstallAddon_Errors(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	t.Run("invalid addon file", func(t *testing.T) {
		_, err := InstallAddonInServer(getAddonFixturePath(t, "corrupt.mcaddon"), tempServerDir)
		require.Error(t, err)
	})

	t.Run("addon with no manifests", func(t *testing.T) {
		_, err := InstallAddonInServer(getAddonFixturePath(t, "no_manifest.zip"), tempServerDir)
		require.Error(t, err)
	})

	t.Run("invalid server directory", func(t *testing.T) {
		newInvalidServerDir := getServerFixturePath(t, "not_a_server")
		_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), newInvalidServerDir)
		require.Error(t, err)
	})

	t.Run("server with level-name set but worlds dir not yet created should fail", func(t *testing.T) {
		// InstallAddonInServer should create directory `worlds/<level-name>/behavior_packs` or `worlds/<level-name>/resource_packs`
		// on the fly even if they don't exist yet.
		tempServerDir := copyServerFixture(t, "not_a_server_missing_worlds")
		defer os.RemoveAll(tempServerDir)

		_ /*installedPacks*/, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
		require.Error(t, err)
	})
}
