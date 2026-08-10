package libminecraftbedrock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestServer_IsValid(t *testing.T) {
	cases := []struct {
		name      string
		fixture   string
		wantValid bool
	}{
		{"valid server", "server", true},
		{"valid server with installed pack", "server_with_installed_pack", true},
		{"missing executable", "not_a_server_missing_exec", false},
		{"missing server.properties", "not_a_server_missing_server.properties", false},
		{"missing worlds directory", "not_a_server_missing_worlds", false},
		{"not a server at all", "not_a_server", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			server := Server{
				Path: getServerFixturePath(t, tc.fixture),
			}
			require.Equal(t, tc.wantValid, server.IsValid())
		})
	}
}

func TestServer_ActiveWorld(t *testing.T) {
	t.Run("valid server resolves its active world", func(t *testing.T) {
		server := Server{
			Path: getServerFixturePath(t, "server"),
		}
		world, err := server.ActiveWorld()
		require.NoError(t, err)
		require.NotNil(t, world)
		require.True(t, fileExists(world.Path))
	})

	t.Run("invalid server returns an error", func(t *testing.T) {
		server := Server{Path: getServerFixturePath(t, "not_a_server")}
		world, err := server.ActiveWorld()
		require.Error(t, err)
		require.Nil(t, world)
	})
}

func TestGetServer(t *testing.T) {
	t.Run("valid server directory", func(t *testing.T) {
		path := getServerFixturePath(t, "server")
		server, err := GetServer(path)
		require.NoError(t, err)
		require.NotNil(t, server)
	})

	t.Run("invalid server directory", func(t *testing.T) {
		path := getServerFixturePath(t, "not_a_server")
		server, err := GetServer(path)
		require.Error(t, err)
		require.Nil(t, server)
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		server, err := GetServer("/path/does/not/exist/at/all")
		require.Error(t, err)
		require.Nil(t, server)
	})
}

/*
func TestBuildServerWithAllPacks(t *testing.T) {
	server := Server{Path: getServerFixturePath(t, "server_with_installed_packs")}
	world, err := server.ActiveWorld()
	require.NoError(t, err)

	world.InstallAddon(getAddonFixturePath(t, "behavior_only.mcpack"))
	world.InstallAddon(getAddonFixturePath(t, "foobar.mcaddon"))
	world.InstallAddon(getAddonFixturePath(t, "solo.mcpack"))
}
*/
