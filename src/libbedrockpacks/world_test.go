package libminecraftbedrock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- FindActiveWorldDir / FindWorldDirectories (package-level helpers used by Server.ActiveWorld) ---

func TestFindActiveWorldDir(t *testing.T) {
	cases := []struct {
		testName     string
		fixture      string
		expectSuffix string
		expectError  bool
	}{
		{"Test random directory", "not_a_server", "", true},
		{"Test missing bedrock_server.exe", "not_a_server_missing_exec", "", true},
		{"Test no server.properties", "not_a_server_missing_server.properties", "", true},
		{"Test no worlds directory", "not_a_server_missing_worlds", "", true},

		{"Test server with content", "server", "worlds/Bedrock level", false},
		{"Test missing level-name falls back to first world dir", "server_no_level_name", "worlds/MyWorld", false},
		{"Test full server", "server_with_installed_pack", "worlds/Bedrock level", false},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			serverDir := getServerFixturePath(t, tc.fixture)
			got, err := FindActiveWorldDir(serverDir)

			if !tc.expectError {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			if !tc.expectError {
				// FindActiveWorldDir() returns absolute paths
				// Build an absolute path from the expected relative path

				want := serverDir + "/" + tc.expectSuffix
				want = filepath.Clean(want) // normalize expected path to match the file separator on the system
				require.Equal(t, want, got)
			}
		})
	}
}

func TestFindWorldDirectories(t *testing.T) {
	t.Run("server with a single world", func(t *testing.T) {
		serverDir := getServerFixturePath(t, "server")
		got, err := FindWorldDirectories(serverDir)
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.Equal(t, filepath.Join(serverDir, "worlds", "Bedrock level"), got[0])
	})

	t.Run("server missing the worlds directory", func(t *testing.T) {
		serverDir := getServerFixturePath(t, "not_a_server_missing_worlds")
		worlds, err := FindWorldDirectories(serverDir)
		require.Error(t, err)
		require.Nil(t, worlds)
	})

	t.Run("invalid server directory", func(t *testing.T) {
		serverDir := getServerFixturePath(t, "not_a_server")
		_, err := FindWorldDirectories(serverDir)
		require.Error(t, err)
	})
}

func TestWorld_Name(t *testing.T) {
	t.Run("falls back to directory base name when levelname.txt is missing", func(t *testing.T) {
		world := getTempActiveWorld(t, "server")
		name, err := world.Name()
		require.NoError(t, err)
		require.Equal(t, filepath.Base(world.Path), name)
	})

	t.Run("uses levelname.txt content when present", func(t *testing.T) {
		world := getTempActiveWorld(t, "server")

		// Create a "levelname.txt" file in world directory
		err := os.WriteFile(filepath.Join(world.Path, "levelname.txt"), []byte("My Cool World"), 0o644)
		require.NoError(t, err)

		name, err := world.Name()
		require.NoError(t, err)

		require.Equal(t, "My Cool World", name)
	})
}

func TestWorld_PacksInstallDir(t *testing.T) {
	world := getTempActiveWorld(t, "server_with_multiple_packs")

	dir1, err := world.PacksInstallDir(BehaviorPack)
	require.NoError(t, err)

	dir2, err := world.PacksInstallDir(ResourcePack)
	require.NoError(t, err)

	require.NotEqual(t, dir1, dir2)

	_, err = world.PacksInstallDir(UnknownPack)
	require.Error(t, err)
}

func TestWorld_PacksByKind(t *testing.T) {
	world := getTempActiveWorld(t, "server_with_multiple_packs")

	behaviorPacks, err := world.PacksByKind(BehaviorPack)
	require.NoError(t, err)
	require.Len(t, behaviorPacks, 2)

	require.Equal(t, "Solo BP", behaviorPacks[0].Name())
	require.Equal(t, "Foobar BP", behaviorPacks[1].Name())

	resourcePacks, err := world.PacksByKind(ResourcePack)
	require.NoError(t, err)
	require.Len(t, resourcePacks, 2)

	require.Equal(t, "Solo RP", resourcePacks[0].Name())
	require.Equal(t, "Foobar RP", resourcePacks[1].Name())
}

func TestWorld_Packs(t *testing.T) {
	t.Run("world with a registered pack", func(t *testing.T) {
		world := getTempActiveWorld(t, "server_with_multiple_packs")
		packs, err := world.Packs()
		require.NoError(t, err)

		require.Len(t, packs, 4)

		require.Equal(t, "Solo BP", packs[0].Name())
		require.Equal(t, "Foobar BP", packs[1].Name())
		require.Equal(t, "Solo RP", packs[2].Name())
		require.Equal(t, "Foobar RP", packs[3].Name())
	})

	t.Run("empty world", func(t *testing.T) {
		world := getTempActiveWorld(t, "server")
		packs, err := world.Packs()
		require.NoError(t, err)
		require.Empty(t, packs)
	})
}

func TestWorld_PacksByUUID(t *testing.T) {
	world := getTempActiveWorld(t, "server_with_installed_pack")

	t.Run("known uuid", func(t *testing.T) {
		pack, err := world.PacksByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c")
		require.NoError(t, err)
		require.NotNil(t, pack)
		require.Equal(t, "Foobar BP", pack.Name())
	})

	t.Run("unknown uuid returns nil, no error", func(t *testing.T) {
		pack, err := world.PacksByUUID("00000000-0000-0000-0000-000000000000")
		require.NoError(t, err)
		require.Nil(t, pack)
	})
}

func TestWorld_RegisterUnregisterIsPackRegistered(t *testing.T) {
	world := getTempActiveWorld(t, "server_with_installed_pack")

	// Find a pre-registered pack by UUID
	pack, err := world.PacksByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c")
	require.NoError(t, err)
	require.NotNil(t, pack)

	// Assert already registered
	registered, err := world.IsPackRegistered(pack)
	require.NoError(t, err)
	require.True(t, registered, "assertion failed: pack should be registered")

	// Unregister then assert NOT already registered
	require.NoError(t, world.UnregisterPack(pack))
	registered, err = world.IsPackRegistered(pack)
	require.NoError(t, err)
	require.False(t, registered, "pack should no longer be registered")

	// Try to register again
	require.NoError(t, world.RegisterPack(pack))
	registered, err = world.IsPackRegistered(pack)
	require.NoError(t, err)
	require.True(t, registered, "pack should be registered again")
}

func TestWorld_InstallAddon(t *testing.T) {
	world := getTempActiveWorld(t, "server")

	installed, err := world.InstallAddon(getAddonFixturePath(t, "foobar.mcaddon"))
	require.NoError(t, err)
	require.Len(t, installed, 2)

	// Assert the packs in addon are installed in the world
	_, err = world.PacksByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c") // foobar.mcaddon/foobar_BP
	require.NoError(t, err)

	_, err = world.PacksByUUID("33333333-3333-3333-3333-333333333333") // foobar.mcaddon/foobar_RP
	require.NoError(t, err)

	// Assert both are registered
	for _, pack := range installed {
		registered, err := world.IsPackRegistered(pack)
		require.NoError(t, err)
		require.True(t, registered)
	}
}

func TestWorld_InstallPack(t *testing.T) {
	world := getTempActiveWorld(t, "server")

	// Load a pack (unzip and load) and install it directly through World.InstallPack.
	extractDir := t.TempDir()
	require.NoError(t, ExtractZip(getAddonFixturePath(t, "solo.mcpack"), extractDir))
	pack, err := LoadPackFromDirectory(extractDir)
	require.NoError(t, err)

	// Install
	installedPack, err := world.InstallPack(pack)
	require.NoError(t, err)
	assert.Equal(t, "Solo RP", installedPack.Name())

	// Assert registered
	registered, err := world.IsPackRegistered(installedPack)
	require.NoError(t, err)
	assert.True(t, registered)
}

func TestWorld_UninstallAddon(t *testing.T) {
	world := getTempActiveWorld(t, "server")

	// install addons in world
	installed, err := world.InstallAddon(getAddonFixturePath(t, "foobar.mcaddon"))
	require.NoError(t, err)
	require.Len(t, installed, 2)

	// Uninstall the same addons in world
	uninstalled, err := world.UninstallAddon(getAddonFixturePath(t, "foobar.mcaddon"))
	require.NoError(t, err)
	require.Len(t, uninstalled, 2)

	// Assert no packs should be found in world
	packs, err := world.Packs()
	require.NoError(t, err)
	assert.Empty(t, packs)

	// Assert packs are not found in
	for _, pack := range uninstalled {
		// Assert the path "used" to be specified
		require.NotEmpty(t, pack.Path)

		// Assert directory was deleted
		assertDirNotExists(t, pack.Path)
	}
}

func TestWorld_UninstallPack(t *testing.T) {
	world := getTempActiveWorld(t, "server_with_installed_pack")

	// Get installed Foobar BP pack
	pack, err := world.PacksByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c")
	require.NoError(t, err)
	require.NotNil(t, pack)

	// Uninstall
	uninstalledPack, err := world.UninstallPack(pack)
	require.NoError(t, err)
	require.Equal(t, "Foobar BP", uninstalledPack.Name())

	// Assert not registered anymore
	registered, err := world.IsPackRegistered(uninstalledPack)
	require.NoError(t, err)
	require.False(t, registered)

	// Assert the pack's directory does not exists
	assertDirNotExists(t, uninstalledPack.Path)
}
