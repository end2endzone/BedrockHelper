package libbedrockpacks

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInstallAddon_MultiPackAddon(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	// Install the addon
	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)
	require.Len(t, installedPacks, 2)

	// Verify both packs are registered
	server, err := GetServer(tempServerDir)
	world, err := server.ActiveWorld()
	for _, pack := range installedPacks {
		registered, err := world.IsPackRegistered(pack)
		require.NoError(t, err)
		require.True(t, registered)
	}

	// Assert that all packs are in installedPacks
	for _, p := range installedPacks {
		require.True(t, p.Name() == "Foobar BP" || p.Name() == "Foobar RP")
	}
}

func TestInstallAddon_SinglePack(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	// Install the single pack
	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "solo.mcpack"), tempServerDir)
	require.NoError(t, err)
	require.Len(t, installedPacks, 1)

	// Verify the pack is registered
	server, err := GetServer(tempServerDir)
	world, err := server.ActiveWorld()
	for _, pack := range installedPacks {
		registered, err := world.IsPackRegistered(pack)
		require.NoError(t, err)
		require.True(t, registered)
	}

	// Assert that all the pack is in installedPacks
	pack := installedPacks[0]
	require.Equal(t, ResourcePack, pack.KindSafe())
	require.Equal(t, "Solo RP", pack.Name())
}

func TestInstallAddon_ReinstallReplacesExisting(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	// Install once
	_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)

	// Install again
	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err, "reinstall failed")
	require.Len(t, installedPacks, 2)

	// Verify the pack is installed only once
	server, err := GetServer(tempServerDir)
	world, err := server.ActiveWorld()
	packs, err := world.Packs()
	require.NoError(t, err)
	require.Len(t, packs, 1, "expected reinstall to not duplicate the installed pack directory")
}

func TestInstallAddon_Errors(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	t.Run("invalid addon file", func(t *testing.T) {
		_, err := InstallAddonInServer(getAddonFixturePath(t, "corrupted.mcaddon"), tempServerDir)
		require.Error(t, err)
	})

	t.Run("addon with no manifests installs zero packs without an error", func(t *testing.T) {
		// LoadAllPacksFromDirectoriesOrSubdirectories (used by InstallAddon) returns an empty list when no manifest.json is missing.
		// The function should not return an error in such a case.
		installed, err := InstallAddonInServer(getAddonFixturePath(t, "zip_with_no_manifest.zip"), tempServerDir)
		require.NoError(t, err)
		require.Empty(t, installed)
	})

	t.Run("invalid server directory", func(t *testing.T) {
		newInvalidServerDir := getServerFixturePath(t, "not_a_server")
		_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), newInvalidServerDir)
		require.Error(t, err)
	})

	t.Run("server missing the worlds directory", func(t *testing.T) {
		missingWorldsDir := copyServerFixture(t, "not_a_server_missing_worlds")
		_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), missingWorldsDir)
		require.Error(t, err)
	})
}
