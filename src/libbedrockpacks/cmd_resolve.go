package libbedrockpacks

import (
	"fmt"
	"strings"
)

// ResolvePackByUUID searches serverDir (recursively) for an add-on file containing a pack whose manifest UUID matches the given UUID.
// Returns the path of the first matching add-on file found.
// Returns an empty path if no match is found.
// Returns an error otherwise.
func ResolvePackByUUID(uuid string, serverDir string) (string, error) {
	err := ValidateServerDirectory(serverDir)
	if err != nil {
		return "", err
	}

	// Find all addons in a directory
	addonPaths, err := FindAddonsInDir(serverDir, true)
	if err != nil {
		return "", err
	}

	// for each addons
	for _, addonPath := range addonPaths {

		// find all manifests inside the addon
		manifestPaths, err := FindManifestsRelativePathInAddon(addonPath)
		if err != nil {
			continue
		}

		// for each manifest
		for _, mp := range manifestPaths {
			// Get tje manifest json RAW bytes
			data, err := readZipEntry(addonPath, mp)
			if err != nil {
				continue
			}

			// Parse it
			m, err := LoadManifestFromBytes(data)
			if err != nil {
				continue
			}

			// Is that the manifest of the pack we are looking for ?
			if strings.EqualFold(m.Header.UUID, uuid) {
				return addonPath, nil
			}
		}
	}

	return "", fmt.Errorf("no add-on file under %q contains a pack with uuid %s", serverDir, uuid)
}
