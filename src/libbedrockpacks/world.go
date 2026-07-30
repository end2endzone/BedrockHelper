package libbedrockpacks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
