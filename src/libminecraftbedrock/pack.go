package libminecraftbedrock

import (
	"fmt"
	"io/fs"
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
	dirName := sanitizeCharactersInPath(p.Manifest.Header.Name)
	return dirName
}

func (p Pack) UUID() string {
	return p.Manifest.Header.UUID
}

func (p Pack) Description() string {
	safeKind, err := p.Kind()
	if err != nil {
		safeKind = UnknownPack
	}

	desc := fmt.Sprintf("%s (%s) [%s] uuid=%s -> %s\n", p.Name(), safeKind, p.Manifest.Header.Version, p.Manifest.Header.UUID, p.Path)
	return desc
}

// LoadPackFromDirectory loads a pack stored in the given directory.
// The given directory path must contains a manifest.json file to be a valid pack.
// Returns a valid pack or an error otherwise.
func LoadPackFromDirectory(path string) (*Pack, error) {
	manifestFullPath := filepath.Join(path, "manifest.json")

	// Load its manifest
	manifest, err := LoadManifestFromFile(manifestFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read a pack from directory: %q, got error %v", path, err)
	}

	pack := &Pack{
		Path:     path,
		Manifest: manifest,
	}
	return pack, nil
}

// LoadPacksFromSubdirectories browse the sub directories from the given directory and loads a pack from each subdir.
// All sub directories must be a valid pack directory, otherwise the function returns an error.
// Returns a valid list of packs. Returns an empty list if there are no subdirectories.
// Returns an error otherwise.
func LoadPacksFromSubdirectories(path string) ([]*Pack, error) {
	var packs []*Pack

	// Get all the sub directories
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read multiple packs from directory: %q, got error %v", path, err)
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

// LoadAllPacksFromDirectoriesOrSubdirectories recursively traverses the given directory to detect and load directories containing Packs.
// This function is compatible with `.mcpack` and `.mcaddon` files.
// Extension `.mcpack`  have a manifest.json file at the root directory.
// Extension `.mcaddon` have a manifest.json file in each sub directory.
// Returns an error if a directory containing a manifest.json which fails to load as a pack.
// Returns am empty pack list when no manifest.json files is found
func LoadAllPacksFromDirectoriesOrSubdirectories(root string) ([]*Pack, error) {
	var packs []*Pack

	/*
		// Try to load a pack at the root directory
		rootPack, err1 := LoadPackFromDirectory(path)
		if err1 == nil {
			// success for the root directory
			packs = append(packs, rootPack)
			return packs, nil
		}

		// Try to fallback to loading packs from each subdirectories
		packs, err2 := LoadPacksFromSubdirectories(path)
		if err2 == nil {
			// success for the sub directories
			return packs, nil
		}
	*/

	err := filepath.WalkDir(root, func(path string, dir fs.DirEntry, err error) error {
		// If there is an error accessing a path, return it to stop walking
		if err != nil {
			return err
		}

		// Ignore files entirely
		if !dir.IsDir() {
			return nil
		}

		// Does directory contains a manifest.json file ?
		manifestPath := filepath.Join(path, "manifest.json")
		if fileExists(manifestPath) {
			// This directory contains a manifest, load this directory as a pack.
			pack, err := LoadPackFromDirectory(path)
			if err != nil {
				// Pack failed to load
				return err
			}

			packs = append(packs, pack)
		}
		return nil
	})

	if err != nil {
		// An error occured while walking the directories
		return nil, fmt.Errorf("failed to detect packs from directory: %q, got %v", root, err)
	}

	// Success
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
