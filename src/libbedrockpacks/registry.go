package libbedrockpacks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

// registryEntry mirrors one element of world_behavior_packs.json or world_resource_packs.json.
type registryEntry struct {
	PackID  string  `json:"pack_id"`
	Version Version `json:"version"`
}

// LoadRegistryFile reads the world registry file.
// A registry file contains a list of registryEntry.
// If the file does not exist, an empty (not nil-error) list is returned.
func LoadRegistryFile(path string) ([]registryEntry, error) {
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
	err = json.Unmarshal(data, &entries)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q: %w", path, err)
	}

	return entries, nil
}

// WriteRegistryFile write the given registry entries the world registry file.
// A registry file contains a list of registryEntry.
// If the file does not exist, the file is created an empty (not nil-error) list is returned.
func WriteRegistryFile(path string, entries []registryEntry) error {
	// Marshal the entries into an array of bytes
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

// RegisterPackInRegistryFile adds (or updates) a pack entry in the given world registry file.
// Registering an already-registered pack updates the existing entry. It does not create duplicate the entry.
func RegisterPackInRegistryFile(path string, uuid string, version Version) error {
	entries, err := LoadRegistryFile(path)
	if err != nil {
		return err
	}

	// Search for the existing UUID pack, if already registered
	for i, e := range entries {
		if strings.EqualFold(e.PackID, uuid) {
			entries[i].Version = version
			return WriteRegistryFile(path, entries)
		}
	}

	// Pack not already registered, add a new entry
	entries = append(entries, registryEntry{
		PackID:  uuid,
		Version: version,
	})

	// Write the new entries to a registry file
	return WriteRegistryFile(path, entries)
}

// UnregisterPackInRegistryFile removes a pack entry in the given world registry file.
// Does not return an error if the given uuid is not already registered.
// To know if a pack is actually registered, use IsPackRegisteredInRegistryFile().
func UnregisterPackInRegistryFile(path string, uuid string, version Version) error {
	entries, err := LoadRegistryFile(path)
	if err != nil {
		return err
	}

	// Search for the existing UUID pack, if found, remove it from the list
	for i, e := range entries {
		if strings.EqualFold(e.PackID, uuid) {
			// Found. Remove it
			entries = slices.Delete(entries, i, i+1)
			continue
		}
	}

	// Write the new entries to a registry file
	return WriteRegistryFile(path, entries)
}

// IsPackRegisteredInRegistryFile checks if a pack entry is registered in the given world registry file.
// Returns an error if the registry file can not be loaded
func IsPackRegisteredInRegistryFile(path string, uuid string, version Version) (bool, error) {
	entries, err := LoadRegistryFile(path)
	if err != nil {
		return false, err
	}

	// Search for the existing UUID pack, if found, remove it from the list
	for _, e := range entries {
		if strings.EqualFold(e.PackID, uuid) {
			// Found !
			return true, nil
		}
	}

	return false, nil
}

func getRegistryFilePathFromWorldDirAndKind(worldDir string, kind PackKind) (string, error) {
	registryFileName, err := kind.RegistryFileName()
	if err != nil {
		return "", err
	}
	path := filepath.Join(worldDir, registryFileName)
	return path, nil
}
