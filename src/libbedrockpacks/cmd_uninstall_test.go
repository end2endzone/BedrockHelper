package libbedrockpacks

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUninstallAddon(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)

	uninstalledPacks, err := UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 2, len(uninstalledPacks), "expected 2 packs uninstalled")

	// Verify both registry files were deleted.
	worldDir, _ := FindActiveWorldDir(tempServerDir)
	bpEntries, _ := readRegistry(worldDir, BehaviorPack)
	rpEntries, _ := readRegistry(worldDir, ResourcePack)
	require.Equal(t, 0, len(bpEntries))
	require.Equal(t, 0, len(rpEntries))

	// Assert that each uninstalled pack's directory were deleted
	for _, p := range uninstalledPacks {
		_, err := os.Stat(p.Path)
		require.True(t, os.IsNotExist(err), "expected pack directory %q to be removed, stat err = %v", p.Path, err)
	}
}

func TestUninstallAddon_NotInstalled(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	// Never installed, so this should fail to find the pack in the world.
	_, err := UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.Error(t, err)
}

func TestUninstallPackByUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")
	defer os.RemoveAll(tempServerDir)

	pack, err := UninstallPackInServerByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c", tempServerDir)
	require.NoError(t, err)
	require.Equal(t, "Foobar BP", pack.Name())
	require.Equal(t, BehaviorPack, pack.KindSafe())

	worldDir, _ := FindActiveWorldDir(tempServerDir)
	entries, _ := readRegistry(worldDir, BehaviorPack)
	require.Equal(t, 0, len(entries))

	_, err = os.Stat(pack.Path)
	require.True(t, os.IsNotExist(err), "expected pack directory to be removed")
}

func TestUninstallPackByUUID_UnknownUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")
	defer os.RemoveAll(tempServerDir)

	_, err := UninstallPackInServerByUUID("00000000-0000-0000-0000-000000000000", tempServerDir)
	require.Error(t, err)
}

func TestInstallThenUninstallByUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	installed, err := InstallAddonInServer(getAddonFixturePath(t, "behavior_only.mcpack"), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(installed))

	pack, err := UninstallPackInServerByUUID(installed[0].UUID(), tempServerDir)
	require.NoError(t, err)
	require.Equal(t, "Solo BP", pack.Name())
}
