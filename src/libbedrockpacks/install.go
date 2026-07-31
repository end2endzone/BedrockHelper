package libbedrockpacks

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallAddonInServer installs every pack contained in the given add-on file (addonPath)
// into the given Minecraft Bedrock server located at serverDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func InstallAddonInServer(addonPath string, serverDir string) ([]InstalledPack, error) {
	err := ValidateAddonFile(addonPath)
	if err != nil {
		return nil, err
	}

	// Is that a valid server installation ?
	err = ValidateServerDirectory(serverDir)
	if err != nil {
		return nil, err
	}

	// Identify all manifests in the addon
	manifestsPathsInAddon, err := FindManifestsRelativePathInAddon(addonPath)
	if err != nil {
		return nil, err
	}

	// Drop root/header manifests
	targets := filterManifestPaths(manifestsPathsInAddon)
	if len(targets) == 0 {
		return nil, fmt.Errorf("add-on %q does not contain any installable packs", addonPath)
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

	// Find active world directory
	worldDir, err := FindActiveWorldDir(serverDir)
	if err != nil {
		return nil, err
	}

	// Install the packs found in the addon archive, one by one.
	var installed []InstalledPack
	for _, manifestRelPath := range targets {
		manifestFullPath := filepath.Join(tempDir, filepath.FromSlash(manifestRelPath))

		manifest, err := LoadManifestFromFile(manifestFullPath)
		if err != nil {
			return installed, fmt.Errorf("failed to load manifest: %w", err)
		}

		// Identify its kind to know where to install
		kind, err := IdentifyPackKind(manifest)
		if err != nil {
			return installed, fmt.Errorf("could not identify pack kind for %q in %q: %w", manifestRelPath, addonPath, err)
		}

		// Define its installation sub directory within the server's world directory
		kindDirName, err := kind.InstallDirName()
		if err != nil {
			return installed, err
		}

		packSourceDir := filepath.Dir(manifestFullPath)
		targetParent := filepath.Join(worldDir, kindDirName)
		err = os.MkdirAll(targetParent, 0o755)
		if err != nil {
			return installed, fmt.Errorf("failed to create %q: %w", targetParent, err)
		}

		dirName := sanitizeCharactersInPath(manifest.Header.Name)
		packTargetInstallDir := filepath.Join(targetParent, dirName)

		// Replace any existing install of the same pack directory name.
		_, err = os.Stat(packTargetInstallDir)
		if err == nil {
			// packTargetInstallDir already exists
			err := os.RemoveAll(packTargetInstallDir)
			if err != nil {
				return installed, fmt.Errorf("failed to remove existing pack directory %q: %w", packTargetInstallDir, err)
			}
		}

		// Move the unzipped pack to the server world's directory
		err = moveDir(packSourceDir, packTargetInstallDir)
		if err != nil {
			return installed, fmt.Errorf("failed to install pack %q: %w", manifest.Header.Name, err)
		}

		// Build registry file path from worldDir and kind (`world_behavior_packs.json` or `world_resource_packs.json`)
		registryFilePath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, kind)
		if err != nil {
			return nil, err
		}

		// Register the pack in `world_behavior_packs.json` or `world_resource_packs.json`.
		err = RegisterPackInRegistryFile(registryFilePath, manifest.Header.UUID, manifest.Header.Version)
		if err != nil {
			return installed, fmt.Errorf("failed to register pack %q: %w", manifest.Header.Name, err)
		}

		installed = append(installed, InstalledPack{
			UUID:      manifest.Header.UUID,
			Name:      manifest.Header.Name,
			Kind:      kind,
			Version:   manifest.Header.Version,
			Directory: packTargetInstallDir,
		})
	}

	return installed, nil
}

// InstallAddonInWorld installs every pack contained in the given add-on file (addonPath)
// into the given Minecraft Bedrock world directory at worldDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func InstallAddonInWorld(addonPath string, worldDir string) ([]*Pack, error) {
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

	// Install each packs into the world
	var installed []*Pack
	for _, pack := range packs {

		newPack, err := InstallPackInWorld(pack, worldDir)
		if err != nil {
			return nil, err
		}

		// The pack has installed succesfully
		installed = append(installed, newPack)
	}

	return installed, nil
}

// InstallPackInWorld installs the given pack
// into the given Minecraft Bedrock world directory at worldDir.
// During the installation process, the pack's original directory is moved to a new location.
// Returns the pack's updated information if installed succesfully.
// Returns an error otherwise.
func InstallPackInWorld(pack *Pack, worldDir string) (*Pack, error) {
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
	packSourceDir := pack.Path
	packsInstallDir := filepath.Join(worldDir, kindSubDir)
	packTargetDir := filepath.Join(packsInstallDir, filepath.Base(pack.Path))

	// Replace any existing install of the same pack directory name.
	_, err = os.Stat(packTargetDir)
	if err == nil {
		// packTargetDir already exists
		err := os.RemoveAll(packTargetDir)
		if err != nil {
			return nil, fmt.Errorf("failed to delete existing pack directory %q: %w", packTargetDir, err)
		}
	}

	// Check for existing copy or an older version that would be already installed
	existingPacks, err := LoadPacksFromSubdirectories(packsInstallDir)
	if err == nil {
		existingPacks := FilterPacksByUUID(existingPacks, pack.Manifest.Header.UUID)
		for _, existingPack := range existingPacks {
			// One or multiple existing version of this pack is already installed.
			// Uninstall them first

			_ /*uninstalledPack*/, err := UninstallPackInWorld(existingPack, worldDir)
			if err != nil {
				return nil, fmt.Errorf("existing or older pack has failed to uninstall first %q: %w", existingPack.Path, err)
			}
		}
	}

	// Move the input pack directory to the server world's directory
	err = moveDir(packSourceDir, packTargetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to install pack %q: %w", pack.Name(), err)
	}

	// Build registry file path from worldDir and kind (`world_behavior_packs.json` or `world_resource_packs.json`)
	registryFilePath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, kind)
	if err != nil {
		return nil, err
	}

	// Register the pack in the right registry file.
	err = RegisterPackInRegistryFile(registryFilePath, pack.Manifest.Header.UUID, pack.Manifest.Header.Version)
	if err != nil {
		return nil, fmt.Errorf("failed to register pack %q: %w", pack.Manifest.Header.Name, err)
	}

	// Parse the pack after installation to be sure we get its updated properties.
	installedPack, err := LoadPackFromDirectory(packTargetDir)
	if err != nil {
		return nil, err
	}

	return installedPack, nil
}
