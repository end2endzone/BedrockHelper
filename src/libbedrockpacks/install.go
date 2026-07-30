package libbedrockpacks

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallAddon installs every pack contained in the given add-on file (addonPath)
// into the given Minecraft Bedrock server located at serverDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func InstallAddon(addonPath string, serverDir string) ([]InstalledPack, error) {
	// Is that an addon ?
	ok, err := IsValidAddonFile(addonPath)
	if !ok || err != nil {
		return nil, err
	}

	// Is that a valid server installation ?
	ok, err = IsValidServerDirectory(serverDir)
	if !ok || err != nil {
		return nil, fmt.Errorf("not a server installation directory: %v", serverDir)
	}

	// Identify all manifests in the addon
	manifestsPathsInAddon, err := FindManifestsInAddon(addonPath)
	if err != nil {
		return nil, err
	}

	// Drop root/header manifests
	targets := filterManifestPaths(manifestsPathsInAddon)
	if len(targets) == 0 {
		return nil, fmt.Errorf("add-on %q does not contain any installable packs", addonPath)
	}

	// Create a tempoerary directory to unzip the archive
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

		dirName := sanitizePackDirName(manifest.Header.Name)
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

		// Register the pack in `world_behavior_packs.json` or `world_resource_packs.json`.
		err = registerPack(worldDir, kind, manifest.Header.UUID, manifest.Header.Version)
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
