package libminecraftbedrock

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveAddonByUUID(t *testing.T) {
	// Stage a copy of the add-on file inside the server directory so we can resolve it by UUID.
	tempServerDir := copyServerFixture(t, "server_empty")

	// Create a sub directory for new addons in server
	newAddonsDir := filepath.Join(tempServerDir, "new_addons")
	require.NoError(t, os.MkdirAll(newAddonsDir, 0o755))

	// Copy an addon fixture to our temp server directory
	sourceFile := getAddonFixturePath(t, "foobar.mcaddon")
	targetFile := filepath.Join(newAddonsDir, "foobar.mcaddon")
	err := copyFile(sourceFile, targetFile)

	foobarRessourcePackUUID := "33333333-3333-3333-3333-333333333333"

	// now try to resolve the new addon by UUID
	got, err := ResolveAddonByUUID(foobarRessourcePackUUID, tempServerDir)
	require.NoError(t, err)

	// assert
	assert.Equal(t, targetFile, got)
}

func TestResolveAddonByUUID_NotFound(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")

	_, err := ResolveAddonByUUID("00000000-0000-0000-0000-000000000000", tempServerDir)
	assert.Error(t, err)
}

func TestResolveAddonByUUID_InvalidServer(t *testing.T) {
	tempServerDir := copyServerFixture(t, "not_a_server")

	path, err := ResolveAddonByUUID("33333333-3333-3333-3333-333333333333", tempServerDir)
	assert.Error(t, err)
	assert.Empty(t, path)
}
