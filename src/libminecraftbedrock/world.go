package libminecraftbedrock

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type World struct {
	Path string
}

// Name returns the name of the World.
// The name of a level is stored in file `levelname.txt`.
// If this file is missing, the function fall back to using the directory name as the name of the world.
func (w World) Name() (string, error) {
	// Check for a levelname.txt
	levelnamePath := filepath.Join(w.Path, "levelname.txt")
	info, err := os.Stat(levelnamePath)
	if err == nil && !info.IsDir() {
		// File found!

		// Read file content
		data, err := os.ReadFile(levelnamePath)
		if err == nil {
			// Parse as UTF-8 string
			name := string(data[:])
			return name, nil
		}
	}

	// Fall back to base directory name
	name := filepath.Base(w.Path)
	return name, nil
}

// PacksInstallDir provides the installation directory path for a given kind of packs.
func (w World) PacksInstallDir(kind PackKind) (string, error) {
	kindInstallDirName, err := kind.InstallDirName()
	if err != nil {
		return "", err
	}

	// Validate the output directory
	packsDir := filepath.Join(w.Path, kindInstallDirName)
	err = ValidateDirectory(packsDir)
	if err != nil {
		return "", err
	}

	return packsDir, nil
}

// PacksByKind returns the packs installed in this world that matches the given kind of pack.
func (w World) PacksByKind(kind PackKind) ([]*Pack, error) {
	var packs []*Pack

	packsDir, err := w.PacksInstallDir(kind)
	if err != nil {
		return packs, err
	}

	// Search all directories under this kind
	packs, err = LoadPacksFromSubdirectories(packsDir)
	if err != nil {
		return packs, err
	}

	return packs, nil
}

// Packs returns the packs installed in this world.
func (w World) Packs() ([]*Pack, error) {
	var packs []*Pack

	// For each kind
	for _, kind := range AllPackKinds {

		// Get all the packs of this kind
		newPacks, err := w.PacksByKind(kind)
		if err != nil {
			continue // continue to prevent failing if a "kind" directory is not found in this world
		}

		// keep these packs in the slice
		packs = append(packs, newPacks...)
	}

	return packs, nil
}

// PacksByKind searches the packs installed in this world and returns the pack that matches the given UUID.
// Returns nil if the pack is not found (no error)
func (w World) PacksByUUID(uuid string) (*Pack, error) {
	var packs []*Pack

	packs, err := w.Packs()
	if err != nil {
		return nil, err
	}

	// Is this pack our UUID target ?
	for _, pack := range packs {
		if strings.EqualFold(pack.Manifest.Header.UUID, uuid) {
			// We found a match
			return pack, nil
		}
	}

	return nil, nil
}

// IsPackRegistered checks if the given pack is registered in the world.
// Returns an error if the registry file can not be loaded
func (w World) IsPackRegistered(pack *Pack) (bool, error) {
	kind, err := pack.Kind()
	if err != nil {
		return false, err
	}

	registryFileName, err := kind.RegistryFileName()
	if err != nil {
		return false, err
	}

	registryFilePath := filepath.Join(w.Path, registryFileName)

	registered, err := IsPackRegisteredInRegistryFile(registryFilePath, pack.Manifest.Header.UUID, pack.Manifest.Header.Version)
	return registered, err
}

// RegisterPack registers the given pack in the world.
func (w World) RegisterPack(pack *Pack) error {
	kind, err := pack.Kind()
	if err != nil {
		return err
	}

	registryFileName, err := kind.RegistryFileName()
	if err != nil {
		return err
	}

	registryFilePath := filepath.Join(w.Path, registryFileName)

	err = RegisterPackInRegistryFile(registryFilePath, pack.Manifest.Header.UUID, pack.Manifest.Header.Version)
	return err
}

// UnregisterPack unregisters the given pack in the world.
func (w World) UnregisterPack(pack *Pack) error {
	kind, err := pack.Kind()
	if err != nil {
		return err
	}

	registryFileName, err := kind.RegistryFileName()
	if err != nil {
		return err
	}

	registryFilePath := filepath.Join(w.Path, registryFileName)

	err = UnregisterPackInRegistryFile(registryFilePath, pack.Manifest.Header.UUID, pack.Manifest.Header.Version)
	return err
}

// InstallAddonInWorld installs every pack contained in the given add-on file (addonPath)
// into the given Minecraft Bedrock world directory at worldDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func (w World) InstallAddon(addonPath string) ([]*Pack, error) {
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
	packs, err := LoadAllPacksFromDirectoriesOrSubdirectories(tempDir)
	if err != nil {
		return nil, err
	}

	// Install each packs into the world
	var installed []*Pack
	for _, pack := range packs {

		newPack, err := w.InstallPack(pack)
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
func (w World) InstallPack(pack *Pack) (*Pack, error) {
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
	packsInstallDir := filepath.Join(w.Path, kindSubDir)
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

			_ /*uninstalledPack*/, err := w.UninstallPack(existingPack)
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
	registryFilePath, err := getRegistryFilePathFromWorldDirAndKind(w.Path, kind)
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

// UninstallAddonInWorld uninstalls every pack contained in the given add-on file (addonPath)
// from the given Minecraft Bedrock world directory at worldDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func (w World) UninstallAddon(addonPath string) ([]*Pack, error) {
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
	packs, err := LoadAllPacksFromDirectoriesOrSubdirectories(tempDir)
	if err != nil {
		return nil, err
	}

	// Uninstall each packs from the world
	var uninstalled []*Pack
	for _, pack := range packs {

		uninstalledPack, err := w.UninstallPack(pack)
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
func (w World) UninstallPack(pack *Pack) (*Pack, error) {
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
	packsInstallDir := filepath.Join(w.Path, kindSubDir)
	packTargetDir := filepath.Join(packsInstallDir, filepath.Base(pack.Path))

	// Build registry file path from worldDir and kind (`world_behavior_packs.json` or `world_resource_packs.json`)
	registryFilePath, err := getRegistryFilePathFromWorldDirAndKind(w.Path, kind)
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

// FindActiveWorldDir identifies the active world directory inside a Minecraft Bedrock server installation directory.
// Returns the path to the active world directory.
// Returns an error otherwise.
func FindActiveWorldDir(serverDir string) (string, error) {
	err := ValidateServerDirectory(serverDir)
	if err != nil {
		return "", err
	}

	worldsDir := filepath.Join(serverDir, "worlds")
	propertiesPath := filepath.Join(serverDir, "server.properties")

	levelName, err := readLevelName(propertiesPath)
	if err == nil && levelName != "" {
		// found a name
		worldDir := filepath.Join(worldsDir, levelName)

		// check the directory exists
		info, err := os.Stat(worldDir)
		if err != nil {
			return "", err
		}
		if !info.IsDir() {
			return "", fmt.Errorf("found level-name '%v' but directory not found: %v", levelName, worldDir)
		}

		// success
		return worldDir, nil
	}

	// Property "level-name" property not found
	// Fall back to the first directory in `worlds` directory.
	worldsDirectories, err := FindWorldDirectories(serverDir)
	if err != nil {
		return "", err
	}

	if len(worldsDirectories) > 0 {
		// Found !
		return worldsDirectories[0], nil
	}

	return "", fmt.Errorf("could not determine active world for %q", serverDir)
}

// FindActiveWorldDir found the world directories inside a Minecraft Bedrock server installation directory.
// Returns an error is the server is missing a `worlds` directory.
// Returns an empty list if the server `worlds` directory is empty.
// Returns the paths of the worlds otherwise.
func FindWorldDirectories(serverDir string) ([]string, error) {
	var matches []string

	err := ValidateServerDirectory(serverDir)
	if err != nil {
		return matches, err
	}

	// Check the `worlds` subdirectory
	worldsSubDir := filepath.Join(serverDir, "worlds")
	info, err := os.Stat(worldsSubDir)
	if err != nil {
		return matches, fmt.Errorf("worlds directory %q is missing or unreadable: %w", worldsSubDir, err)
	}
	if !info.IsDir() {
		return matches, fmt.Errorf("expected a directory: %q", worldsSubDir)
	}

	// Get all the world directories
	entries, err := os.ReadDir(worldsSubDir)
	if err != nil {
		return matches, fmt.Errorf("failed to read world directories: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() {
			worldDir := filepath.Join(worldsSubDir, e.Name())
			matches = append(matches, worldDir)
		}
	}

	return matches, nil
}

// readLevelName reads the "level-name" property in a properties file.
// Returns an empty string if the file exists but the property is not set.
// Return the key value if found.
// Returns an error otherwise.
func readLevelName(propsPath string) (string, error) {
	f, err := os.Open(propsPath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			// skip commented lines
			continue
		}

		if !strings.HasPrefix(line, "level-name=") {
			// not our searched key
			continue
		}

		value := strings.TrimSpace(strings.TrimPrefix(line, "level-name="))
		return value, nil
	}
	err = scanner.Err()
	if err != nil {
		return "", err
	}

	// not found
	return "", nil
}
