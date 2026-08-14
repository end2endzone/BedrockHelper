package libminecraftbedrock

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistryRoundTrip(t *testing.T) {
	worldDir := t.TempDir()
	registryPath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, BehaviorPack)
	require.NoError(t, err)

	// Reading a registry that doesn't exist yet returns an empty list, not an error.
	entries, err := LoadRegistryFile(registryPath)
	require.NoError(t, err)
	assert.Empty(t, entries)

	// Test registering custom UUID
	uuid1 := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	uuid2 := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	require.NoError(t, RegisterPackInRegistryFile(registryPath, uuid1, Version{1, 0, 0}))
	require.NoError(t, RegisterPackInRegistryFile(registryPath, uuid2, Version{2, 0, 0}))

	// Load the file with the new registered UUID
	entries, err = LoadRegistryFile(registryPath)
	require.NoError(t, err)
	require.Len(t, entries, 2)

	// Registering the same UUID again should update its version, not duplicate the entry.
	require.NoError(t, RegisterPackInRegistryFile(registryPath, uuid1, Version{1, 5, 0}))
	entries, err = LoadRegistryFile(registryPath)
	require.NoError(t, err)
	require.Len(t, entries, 2, "re-registering should update in place, not duplicate")

	// Find the updated registered pack, check for the new version
	found := false
	for _, e := range entries {
		if e.PackID == uuid1 {
			found = true
			assert.Equal(t, Version{1, 5, 0}, e.Version)
		}
	}
	assert.True(t, found, "expected uuid1 to still be present after update")

	// Unnistall
	require.NoError(t, UnregisterPackInRegistryFile(registryPath, uuid1, Version{1, 5, 0}))

	entries, err = LoadRegistryFile(registryPath)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	assert.Equal(t, uuid2, entries[0].PackID)

	// Unregistering an already-removed uuid is a no-op, not an error.
	assert.NoError(t, UnregisterPackInRegistryFile(registryPath, uuid1, Version{1, 5, 0}))
}

func TestIsPackRegisteredInRegistryFile(t *testing.T) {
	worldDir := t.TempDir()
	registryPath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, BehaviorPack)
	require.NoError(t, err)

	uuid := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"

	// Test an non-existing UUID
	registered, err := IsPackRegisteredInRegistryFile(registryPath, uuid, Version{1, 0, 0})
	require.NoError(t, err)
	assert.False(t, registered)

	// Add an entry
	require.NoError(t, RegisterPackInRegistryFile(registryPath, uuid, Version{1, 0, 0}))

	// Test again while registered
	registered, err = IsPackRegisteredInRegistryFile(registryPath, uuid, Version{1, 0, 0})
	require.NoError(t, err)
	assert.True(t, registered)
}

func TestRegistryFileNames(t *testing.T) {
	worldDir := t.TempDir()

	// Function getRegistryFilePathFromWorldDirAndKind() internally calls PackKind::RegistryFileName()
	behaviorRegistryPath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, BehaviorPack)
	require.NoError(t, err)
	resourceRegistryPath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, ResourcePack)
	require.NoError(t, err)

	require.NotEmpty(t, behaviorRegistryPath)
	require.NotEmpty(t, resourceRegistryPath)
	require.NotEqual(t, behaviorRegistryPath, resourceRegistryPath)
}
