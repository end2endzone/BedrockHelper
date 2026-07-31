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
		uninstalledPack, err := UninstallPackInWorld(pack, worldDir)
		if err != nil {
			return nil, err
		}

		uninstalled = append(uninstalled, *PackToInstalledPack(uninstalledPack))
	}

	return uninstalled, nil
}

// UninstallPackByUUID uninstalls a single pack, identified by a UUID
// from the Minecraft Bedrock server installed at serverDir.
// This function is useful when the original add-on file is no longer available or has been deleted.
func UninstallPackByUUID(uuid, serverDir string) (InstalledPack, error) {
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

	// Uninstall each packs from the world
	var uninstalled []*Pack
	for _, pack := range packs {

		uninstalledPack, err := UninstallPackInWorld(pack, worldDir)
		if err != nil {
			return nil, err
		}

		// The pack has installed succesfully
		uninstalled = append(uninstalled, uninstalledPack)
	}

	return uninstalled, nil
}

// UninstallPackInWorld uninstalls the given pack
// from the given Minecraft Bedrock world directory at worldDir.
func UninstallPackInWorld(pack *Pack, worldDir string) (*Pack, error) {
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
	packsInstallDir := filepath.Join(worldDir, kindSubDir)
	packTargetDir := filepath.Join(packsInstallDir, filepath.Base(pack.Path))

	// Build registry file path from worldDir and kind (`world_behavior_packs.json` or `world_resource_packs.json`)
	registryFilePath, err := getRegistryFilePathFromWorldDirAndKind(worldDir, kind)
	if err != nil {
		return nil, err
	}

	// Unregister the pack in the right registry file.
	err = UnregisterPackInRegistryFile(registryFilePath, pack.Manifest.Header.UUID, pack.Manifest.Header.Version)
	if err != nil {
		return nil, err
	}

	// Parse the pack before deletion to be able to return its updated properties.
	uninstalledPack, err := LoadPackFromDirectory(packTargetDir)
	if err != nil {
		return nil, err
	}

	// Proceed with the directory deletion
	err = os.RemoveAll(packTargetDir)
	if err != nil {
		return nil, fmt.Errorf("failed to remove pack directory %q: %w", packTargetDir, err)
	}

	return uninstalledPack, err
}
