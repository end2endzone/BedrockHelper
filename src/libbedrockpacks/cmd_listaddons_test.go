package libbedrockpacks

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListInstalledPacks(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")
	defer os.RemoveAll(tempServerDir)

	packs, err := ListInstalledPacks(tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 1, len(packs))
	require.Equal(t, "2bda6085-9d71-4d8a-9b9f-74e07b30459c", packs[0].UUID())
	require.Equal(t, "Foobar BP", packs[0].Name())
	require.Equal(t, BehaviorPack, packs[0].KindSafe())
}

func TestListInstalledPacks_EmptyServer(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	packs, err := ListInstalledPacks(tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 0, len(packs))
}

func TestListInstalledPacks_AfterInstallAndUninstall(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)

	packs, err := ListInstalledPacks(tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 2, len(packs))

	_, err = UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	require.NoError(t, err)

	packs, err = ListInstalledPacks(tempServerDir)
	require.NoError(t, err)
	require.Equal(t, 0, len(packs))
}

func TestListInstalledPacks_InvalidServer(t *testing.T) {
	invalidServer := getServerFixturePath(t, "not_a_server")
	_, err := ListInstalledPacks(invalidServer)
	require.Error(t, err)
}
