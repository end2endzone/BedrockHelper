package libbedrockpacks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Pack struct {
	Path     string
	Manifest *AddonManifest
}

func (p Pack) Kind() (PackKind, error) {
	kind, err := IdentifyPackKind(p.Manifest)
	if err != nil {
		return UnknownPack, err
	}

	return kind, nil
}

func (p Pack) Name() string {
	return p.Manifest.Header.Name
}

func (p Pack) NameSanitized() string {
	dirName := sanitizePackDirName(p.Manifest.Header.Name)
	return dirName
}

func (p Pack) Description() string {
	safeKind, err := p.Kind()
	if err != nil {
		safeKind = UnknownPack
	}

	desc := fmt.Sprintf("%s (%s) [%s] uuid=%s -> %s\n", p.Name(), safeKind, p.Manifest.Header.Version, p.Manifest.Header.UUID, p.Path)
	return desc
}

// ListInstalledPacks lists the packs currently registered for a given Minecraft Bedrock server located at serverDir.
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
	for _, kind := range AllPackKinds {
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

func LoadPackFromDirectory(path string) (*Pack, error) {
	manifestFullPath := filepath.Join(path, "manifest.json")

	// Load its manifest
	manifest, err := LoadManifestFromFile(manifestFullPath)
	if err != nil {
		return nil, err
	}

	pack := &Pack{
		Path:     path,
		Manifest: manifest,
	}
	return pack, nil
}

func LoadPacksFromSubdirectories(path string) ([]*Pack, error) {
	var packs []*Pack

	// Get all the sub directories
	entries, err := os.ReadDir(path)
	if err != nil {
		return packs, fmt.Errorf("failed to read packs from directory: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			// Parse pack at this directory
			packDir := filepath.Join(path, e.Name())
			pack, err := LoadPackFromDirectory(packDir)
			if err != nil {
				return packs, err
			}

			packs = append(packs, pack)
		}
	}

	return packs, nil
}
