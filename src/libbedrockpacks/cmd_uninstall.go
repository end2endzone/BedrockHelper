package libbedrockpacks

import (
	"fmt"
	"os"
	"path/filepath"
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

	// Get the server at serverDir
	server, err := GetServer(serverDir)
	if err != nil {
		return nil, err
	}

	// Get the active world
	activeWorld, err := server.ActiveWorld()
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

	// Find packs installed in world.
	installedPacks, err := findInstalledPacksInWorld(worldDir)
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

		packUUID := manifest.Header.UUID

		// Search the pack's UUID in installed pack list
		pack := FindPackByUUID(installedPacks, packUUID)

		// Uninstall this pack
		uninstalledPack, err := activeWorld.UninstallPack(pack)
		if err != nil {
			return nil, err
		}

		uninstalled = append(uninstalled, *PackToInstalledPack(uninstalledPack))
	}

	return uninstalled, nil
}

// UninstallPackInServerByUUID uninstalls a single pack, identified by a UUID
// from the Minecraft Bedrock server installed at serverDir.
// This function is useful when the original add-on file is no longer available or has been deleted.
func UninstallPackInServerByUUID(uuid, serverDir string) (InstalledPack, error) {
	server, err := GetServer(serverDir)
	if err != nil {
		return InstalledPack{}, err
	}

	// Get the active world
	activeWorld, err := server.ActiveWorld()
	if err != nil {
		return InstalledPack{}, err
	}

	// Find all packs in world
	packs, err := activeWorld.Packs()
	if err != nil {
		return InstalledPack{}, err
	}

	// Find the specific pack to uninstall
	pack := FindPackByUUID(packs, uuid)

	// Do the uninstall
	uninstalledPack, err := activeWorld.UninstallPack(pack)

	return *PackToInstalledPack(uninstalledPack), nil
}

func findInstalledPacksInWorld(worldDir string) ([]*Pack, error) {
	// This is a hack !!!
	// This function should be eventually deleted

	info, err := os.Stat(worldDir)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("directory is not a world directory: %q", worldDir)
	}

	w := World{
		Path: worldDir,
	}

	return w.Packs()
}
