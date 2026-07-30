package libbedrockpacks

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

func (w World) Name() (string, error) {
	name, err := GetWorldName(w.Path)
	return name, err
}

func (w World) PacksInstallDir(kind PackKind) (string, error) {
	dir, err := GetWorldPacksInstallDir(w.Path, kind)
	return dir, err
}

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

func (w World) Packs() ([]*Pack, error) {
	var packs []*Pack

	// For each kind
	for _, kind := range AllPackKinds {
		// Get all the packs of this kind
		newPacks, err := w.PacksByKind(kind)
		if err != nil {
			return packs, err
		}

		// keep these packs in the slice
		packs = append(packs, newPacks...)
	}

	return packs, nil
}

func (w World) RegisterPack(pack Pack) error {
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

func (w World) UnregisterPack(pack Pack) error {
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

func (w World) InstallAddon(addonPath string) ([]*Pack, error) {
	packs, err := InstallAddonInWorld(addonPath, w.Path)
	return packs, err
}

func (w World) UninstallAddon(addonPath string) ([]*Pack, error) {
	packs, err := UninstallAddonInWorld(addonPath, w.Path)
	return packs, err
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

// GetWorldName get the name of a given world
func GetWorldName(path string) (string, error) {
	// Validate the given world directory
	info, err := os.Stat(path)
	if err != nil {
		return "", fmt.Errorf("directory %q is not found: %w", path, err)
	}
	if !info.IsDir() {
		return "", fmt.Errorf("not a directory: %q", path)
	}

	// Check for a levelname.txt
	levelnamePath := filepath.Join(path, "levelname.txt")
	info, err = os.Stat(levelnamePath)
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
	name := filepath.Base(path)
	return name, nil
}

// GetWorldPacksInstallDir returns a world's packs installation directory for a given kind of pack
func GetWorldPacksInstallDir(path string, kind PackKind) (string, error) {
	worldDir := path

	kindInstallDirName, err := kind.InstallDirName()
	if err != nil {
		return "", err
	}

	packsDir := filepath.Join(worldDir, kindInstallDirName)

	// Validate the output directory
	err = ValidateDirectory(packsDir)
	if err != nil {
		return "", err
	}

	return packsDir, nil
}
