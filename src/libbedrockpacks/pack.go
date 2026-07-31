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

func (p Pack) KindSafe() PackKind {
	kind, err := IdentifyPackKind(p.Manifest)
	if err != nil {
		return UnknownPack
	}

	return kind
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
	server, err := GetServer(serverDir)
	if err != nil {
		return nil, err
	}

	activeWorld, err := server.ActiveWorld()
	if err != nil {
		return nil, err
	}

	packs, err := activeWorld.Packs()
	if err != nil {
		return nil, err
	}

	// For each pack, resolve its properties (name, uuid, etc)
	var results []RegisteredPack
	for _, pack := range packs {

		registeredPack := RegisteredPack{
			UUID:    pack.Manifest.Header.UUID,
			Name:    pack.Name(),
			Kind:    pack.KindSafe(),
			Version: pack.Manifest.Header.Version,
		}

		// add to our results
		results = append(results, registeredPack)
	}

	return results, nil
}

/*
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
*/

// LoadPackFromDirectory loads a pack stored in the given directory.
// The given directory path must contains a manifest.json file to be a valid pack.
// Returns a valid pack or an error otherwise.
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

// LoadPacksFromSubdirectories browse the subdirectories of the given directory.
// For each subdir found, it tries to load a pack from this subdir.
// All sub directories must be a valid pack directory, otherwise the function returns an error.
// Returns a valid list of packs. Returns an empty list if there are no subdirectories.
// Returns an error otherwise.
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

// FindPackByUUID searches a given list of packs for a pack with the given UUID.
func FindPackByUUID(packs []*Pack, uuid string) *Pack {
	for _, pack := range packs {
		if strings.EqualFold(pack.Manifest.Header.UUID, uuid) {
			// This is the pack we are looking for
			return pack
		}
	}
	return nil
}

// FilterPacksByUUID filters a given list of packs by UUID.
// There should not be multiple packs with the same UUID in the same list.
// This function is mostly for cleanup and integrity.
func FilterPacksByUUID(packs []*Pack, uuid string) []*Pack {
	var results []*Pack
	for _, pack := range packs {
		if strings.EqualFold(pack.Manifest.Header.UUID, uuid) {
			// Match !
			results = append(results, pack)
		}
	}
	return results
}

// FilterPacksByKind filters a given list of packs by PackKind.
func FilterPacksByKind(packs []*Pack, kind PackKind) []*Pack {
	var results []*Pack
	for _, pack := range packs {
		if pack.KindSafe() == kind {
			// Match !
			results = append(results, pack)
		}
	}
	return results
}
