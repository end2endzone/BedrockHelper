package libbedrockpacks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolvePackByUUID(t *testing.T) {
	// This test copy an add-on file inside the server directory so that we can resolve from a UUID.

	// Create a tempoerary copy of the 'server' testdata server.
	// Defer removal of the temporary directory when the function returns
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	// Copy an add-on file inside the server directory so that we can resolve from a UUID.
	src, err := os.ReadFile(getAddonFixturePath(t, "foobar.mcaddon"))
	require.NoError(t, err)

	newServerAddonsSubDir := filepath.Join(tempServerDir, "new_incoming")
	err = os.MkdirAll(newServerAddonsSubDir, 0o755)
	require.NoError(t, err)

	newAddonPath := filepath.Join(newServerAddonsSubDir, "foobar.mcaddon")
	err = os.WriteFile(newAddonPath, src, 0o644)
	require.NoError(t, err)

	foobarMcAddonRessourcePackUUID := "33333333-3333-3333-3333-333333333333"

	// Try to resolve
	got, err := ResolvePackByUUID(foobarMcAddonRessourcePackUUID, tempServerDir)
	require.NoError(t, err)

	// Assert we found our pack inside our addon
	require.Equal(t, newAddonPath, got)
}

func TestResolvePackByUUID_NotFound(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	_, err := ResolvePackByUUID("00000000-0000-0000-0000-000000000000", tempServerDir)
	require.Error(t, err)
}

func TestResolvePackByUUID_InvalidServer(t *testing.T) {
	tempServerDir := copyServerFixture(t, "not_a_server")
	defer os.RemoveAll(tempServerDir)

	path, err := ResolvePackByUUID("33333333-3333-3333-3333-333333333333", tempServerDir)
	require.Error(t, err)
	require.Equal(t, "", path)
}
