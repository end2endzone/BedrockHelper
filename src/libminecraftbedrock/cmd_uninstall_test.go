package libminecraftbedrock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUninstallAddon(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_empty")

	// First install
	installedPacks, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err, "setup install failed")
	require.Len(t, installedPacks, 2)

	// Then uninstall
	uninstalledPacks, err := UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)
	require.Len(t, uninstalledPacks, 2)

	// Get installed packs
	server, err := GetServer(tempServerDir)
	world, err := server.ActiveWorld()
	remaining, err := world.Packs()
	require.NoError(t, err)
	require.Empty(t, remaining, "expected no packs left installed after uninstall")

	// Assert that each uninstalled pack's directory was deleted.
	for _, p := range uninstalledPacks {
		assertDirNotExists(t, p.Path)
	}
}

func TestUninstallAddon_NotInstalled(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_empty")

	// Never installed, so this should fail to find the pack in the world.
	_, err := UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.Error(t, err)
}

func TestUninstallPackByUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")

	// Uninstall with UUID
	pack, err := UninstallPackInServerByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c", tempServerDir)
	require.NoError(t, err)
	require.Equal(t, "Foobar BP", pack.Name())
	require.Equal(t, BehaviorPack, pack.KindSafe())

	// Get installed packs
	server, err := GetServer(tempServerDir)
	world, err := server.ActiveWorld()
	registered, err := world.IsPackRegistered(pack)
	require.NoError(t, err)
	require.False(t, registered)

	assertDirNotExists(t, pack.Path)
}

func TestUninstallPackByUUID_UnknownUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")

	_, err := UninstallPackInServerByUUID("00000000-0000-0000-0000-000000000000", tempServerDir)
	require.Error(t, err)
}

func TestInstallThenUninstallByUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")

	installed, err := InstallAddonInServer(getAddonFixturePath(t, "behavior_only.mcpack"), tempServerDir)
	require.NoError(t, err)
	require.Len(t, installed, 1)

	pack, err := UninstallPackInServerByUUID(installed[0].UUID(), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, "Solo BP", pack.Name())
}
