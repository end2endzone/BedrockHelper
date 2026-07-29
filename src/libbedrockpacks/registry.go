package libbedrockpacks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// registryEntry mirrors one element of world_behavior_packs.json or world_resource_packs.json.
type registryEntry struct {
	PackID  string  `json:"pack_id"`
	Version Version `json:"version"`
}

// readRegistry reads the world registry file for the given pack kind.
// A registry file contains a list of registryEntry.
// If the file does not exist, an empty (not nil-error) list is returned.
func readRegistry(worldDir string, kind PackKind) ([]registryEntry, error) {
	registryFileName, err := kind.RegistryFileName()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(worldDir, registryFileName)

	// Read the registry file
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []registryEntry{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %w", path, err)
	}

	if len(strings.TrimSpace(string(data))) == 0 {
		return []registryEntry{}, nil
	}

	// Parse all registry entries
	var entries []registryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", path, err)
	}

	return entries, nil
}

// writeRegistry writes the world registry file for the given pack kind.
func writeRegistry(worldDir string, kind PackKind, entries []registryEntry) error {
	registryFileName, err := kind.RegistryFileName()
	if err != nil {
		return err
	}
	err = os.MkdirAll(worldDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create world directory %q: %w", worldDir, err)
	}

	// Marshal the entries into an array of bytes
	path := filepath.Join(worldDir, registryFileName)
	data, err := json.MarshalIndent(entries, "", "\t")
	if err != nil {
		return err
	}
	data = append(data, '\n')

	// Write the bytes to a file
	err = os.WriteFile(path, data, 0o644)
	if err != nil {
		return fmt.Errorf("failed to write %q: %w", path, err)
	}

	return nil
}

// registerPack adds (or updates) a pack entry in the world registry file for the given pack kind.
// Registering an already-registered UUID updates its version.
// It does not create duplicate the entry.
func registerPack(worldDir string, kind PackKind, uuid string, version Version) error {
	entries, err := readRegistry(worldDir, kind)
	if err != nil {
		return err
	}

	// Search for the existing UUID pack, if already registered
	for i, e := range entries {
		if strings.EqualFold(e.PackID, uuid) {
			entries[i].Version = version
			return writeRegistry(worldDir, kind, entries)
		}
	}

	// Pack not already registered, add a new entry
	entries = append(entries, registryEntry{
		PackID:  uuid,
		Version: version,
	})

	// Write the new entries to a registry file
	return writeRegistry(worldDir, kind, entries)
}

// unregisterPack removes a pack entry by UUID in the world registry file for the given pack kind.
// Returns an error if the UUID was not found (expect the UUID must be removed).
func unregisterPack(worldDir string, kind PackKind, uuid string) error {
	entries, err := readRegistry(worldDir, kind)
	if err != nil {
		return err
	}

	// BRowse each entry and copy them to another list, skipping any entry that matches our UUID.
	filteredEntries := entries[:0]
	found := false
	for _, e := range entries {
		if strings.EqualFold(e.PackID, uuid) {
			found = true
			continue
		}
		filteredEntries = append(filteredEntries, e)
	}

	if !found {
		return fmt.Errorf("pack %s is not registered in %s", uuid, getSafeRegistryFileName(kind))
	}

	// Write the new entries to a registry file
	return writeRegistry(worldDir, kind, filteredEntries)
}

func getSafeRegistryFileName(kind PackKind) string {
	name, err := kind.RegistryFileName()
	if err != nil {
		return "registry file"
	}
	return name
}
