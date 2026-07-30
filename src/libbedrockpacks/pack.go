package libbedrockpacks

import (
	"os"
	"path/filepath"
	"strings"
)

// ListInstalledPacks lists the packs currently registered for a given Minecraft Bedrock server located at at serverDir.
// For each registered UUID, it attempts to resolve the name of the pack by scanning the corresponding behavior_packs/ or
// resource_packs/ directories for a matching manifest.json.
func ListInstalledPacks(serverDir string) ([]RegisteredPack, error) {
	err := ValidateServerDirectory(serverDir)
	if err != nil {
		return nil, err
	}

	// Find active world directory
	worldDir, err := FindActiveWorldDir(serverDir)
	if err != nil {
		return nil, err
	}

	// For each registry files
	var results []RegisteredPack
	for _, kind := range []PackKind{BehaviorPack, ResourcePack} {
		registryEntries, err := readRegistry(worldDir, kind)
		if err != nil {
			return nil, err
		}

		uuid2names := getPackNamesByUUIDMap(worldDir, kind)

		// For each registry entry in register files
		for _, e := range registryEntries {
			name := uuid2names[strings.ToLower(e.PackID)]
			if name == "" {
				name = "<unknown - pack files missing>"
			}
			results = append(results, RegisteredPack{
				UUID:    e.PackID,
				Name:    name,
				Kind:    kind,
				Version: e.Version,
			})
		}
	}

	return results, nil
}

// getPackNamesByUUIDMap scans a world's install directory to find all installed packs and
// returns a map of uuid (in lowercase) -> pack display name.
func getPackNamesByUUIDMap(worldDir string, kind PackKind) map[string]string {
	uuid2names := make(map[string]string)

	// Get kind's sub directory
	kindSubdirName, err := kind.InstallDirName()
	if err != nil {
		return uuid2names
	}
	kindDir := filepath.Join(worldDir, kindSubdirName)

	// Get packs for this kind
	packDirs, err := os.ReadDir(kindDir)
	if err != nil {
		return uuid2names
	}

	// For each packs
	for _, e := range packDirs {
		if !e.IsDir() {
			continue
		}

		// Load its manifest
		manifestFullPath := filepath.Join(kindDir, e.Name(), "manifest.json")
		manifest, err := LoadManifestFromFile(manifestFullPath)
		if err != nil {
			continue
		}

		// add its name to the map
		uuid2names[strings.ToLower(manifest.Header.UUID)] = manifest.Header.Name
	}

	return uuid2names
}
