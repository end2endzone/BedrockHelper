package libbedrockpacks

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// UninstallAddonInServer uninstalls every pack contained in the add-on file at the given addonPath
// from the Minecraft Bedrock server installed at serverDir.
// It returns the list of packs that were uninstalled or an error.
func UninstallAddonInServer(addonPath, serverDir string) ([]InstalledPack, error) {
	err := ValidateAddonFile(addonPath)
	if err != nil {
		return nil, err
	}

	err = ValidateServerDirectory(serverDir)
	if err != nil {
		return nil, err
	}

	manifestsPathsInAddon, err := FindManifestsRelativePathInAddon(addonPath)
	if err != nil {
		return nil, err
	}
	targets := filterManifestPaths(manifestsPathsInAddon)
	if len(targets) == 0 {
		return nil, fmt.Errorf("add-on %q does not contain any packs to uninstall", addonPath)
	}

	// Create a tempoerary directory to unzip the archive
	tempDir, err := os.MkdirTemp("", "bedrock_helper_uninstall_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Unzip
	err = ExtractZip(addonPath, tempDir)
	if err != nil {
		return nil, err
	}

	// Find active world directory
	worldDir, err := FindActiveWorldDir(serverDir)
	if err != nil {
		return nil, err
	}

	var uninstalled []InstalledPack
	for _, manifestRelPath := range targets {
		manifestFullPath := filepath.Join(tempDir, filepath.FromSlash(manifestRelPath))

		// Read the unzipped manifest's json file
		manifest, err := LoadManifestFromFile(manifestFullPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load manifest: %w", err)
		}

		pack, err := UninstallPackFromWorldByUUID(worldDir, manifest.Header.UUID)
		if err != nil {
			return nil, err
		}
		uninstalled = append(uninstalled, pack)
	}

	return uninstalled, nil
}

// UninstallPackByUUID uninstalls a single pack, identified by a UUID
// from the Minecraft Bedrock server installed at serverDir.
// This function is useful when the original add-on file is no longer available or has been deleted.
func UninstallPackByUUID(uuid, serverDir string) (InstalledPack, error) {
	err := ValidateServerDirectory(serverDir)
	if err != nil {
		return InstalledPack{}, err
	}

	// Find active world directory
	worldDir, err := FindActiveWorldDir(serverDir)
	if err != nil {
		return InstalledPack{}, err
	}

	return UninstallPackFromWorldByUUID(worldDir, uuid)
}

// UninstallPackFromWorldByUUID uninstalls a single pack, identified by a UUID inside worldDir,
// removes its install directory
// and unregisters it from the appropriate world registry file.
func UninstallPackFromWorldByUUID(worldDir, uuid string) (InstalledPack, error) {
	packInstallDir, kind, err := findPackInstallDirByUUID(worldDir, uuid)
	if err != nil {
		return InstalledPack{}, err
	}

	// Find the name and version of the pack
	name := filepath.Base(packInstallDir) // default to the installation directory name
	var version Version

	// Read the manifest's json file
	manifestFullPath := filepath.Join(packInstallDir, "manifest.json")
	manifest, err := LoadManifestFromFile(manifestFullPath)
	if err != nil {
		// Do not fail if we can't read the manifest's true name/version
		//return InstalledPack{}, fmt.Errorf("failed to load manifest: %w", err)
	} else {
		name = manifest.Header.Name
		version = manifest.Header.Version
	}

	// Uninstall from the registry
	err = unregisterPack(worldDir, kind, uuid)
	if err != nil {
		return InstalledPack{}, err
	}

	// Proceed with the directory deletion
	err = os.RemoveAll(packInstallDir)
	if err != nil {
		return InstalledPack{}, fmt.Errorf("failed to remove pack directory %q: %w", packInstallDir, err)
	}

	return InstalledPack{
		UUID:      uuid,
		Name:      name,
		Kind:      kind,
		Version:   version,
		Directory: packInstallDir,
	}, nil
}

// findPackInstallDirByUUID searches a world's behavior_packs and resource_packs directories for an installed pack
// whose manifest.json header.uuid matches the given uuid.
// Returns the pack's install directory and kind.
func findPackInstallDirByUUID(worldDir, uuid string) (string, PackKind, error) {
	// For all kinds
	for _, kind := range AllPackKinds {
		// Build the pack's kind's installation directory
		subdirName, err := kind.InstallDirName()
		if err != nil {
			continue
		}
		parent := filepath.Join(worldDir, subdirName)

		// Find all directories in the pack's installation directory
		entries, err := os.ReadDir(parent)
		if err != nil {
			continue
		}

		// For all directories (a.k.a packs installed)
		for _, e := range entries {
			if !e.IsDir() {
				// Skip files
				continue
			}

			manifestPath := filepath.Join(parent, e.Name(), "manifest.json")

			manifest, err := LoadManifestFromFile(manifestPath)
			if err != nil {
				// No manifest.json or invalid file. Not a valid pack.
				continue
			}

			// Is this pack our UUID target to delete
			if strings.EqualFold(manifest.Header.UUID, uuid) {
				return filepath.Join(parent, e.Name()), kind, nil
			}
		}
	}
	return "", UnknownPack, fmt.Errorf("no installed pack with uuid %s was found under %q", uuid, worldDir)
}

// UninstallAddonInWorld uninstalls every pack contained in the given add-on file (addonPath)
// from the given Minecraft Bedrock world directory at worldDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func UninstallAddonInWorld(addonPath string, worldDir string) ([]*Pack, error) {
	err := ValidateAddonFile(addonPath)
	if err != nil {
		return nil, err
	}

	// Create a temporary directory to unzip the archive
	tempDir, err := os.MkdirTemp("", "bedrock_helper_install_*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary extraction directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Unzip
	err = ExtractZip(addonPath, tempDir)
	if err != nil {
		return nil, err
	}

	// Load packs from the extracted archive
	packs, err := LoadPacksFromSubdirectories(tempDir)
	if err != nil {
		return nil, err
	}

	// Uninstall each packs into the world
	var uninstalled []*Pack
	for _, pack := range packs {

		// Get the pack's kind
		kind, err := pack.Kind()
		if err != nil {
			return nil, err
		}

		// Get installation sub directory based on kind
		kindSubDir, err := kind.InstallDirName()
		if err != nil {
			return nil, err
		}

		// Get the absolute path
		//packSourceDir := pack.Path
		packInstallDir := filepath.Join(worldDir, kindSubDir)
		packTargetDir := filepath.Join(packInstallDir, filepath.Base(pack.Path))

		uninstallTargetPack, err := LoadPackFromDirectory(packTargetDir)
		if err != nil {
			return nil, err
		}

		_, err = UninstallPackFromWorldByUUID(worldDir, uninstallTargetPack.Path)
		if err != nil {
			return nil, err
		}

		// The pack has installed succesfully
		uninstalled = append(uninstalled, uninstallTargetPack)
	}

	return uninstalled, nil
}
