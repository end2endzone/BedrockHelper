package libbedrockpacks

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FindActiveWorldDir identifies the active world directory inside a Minecraft Bedrock server installation directory.
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

	// "level-name" property not found, check for a single directory under `worlds` directory
	entries, err := os.ReadDir(worldsDir)
	if err != nil {
		return "", fmt.Errorf("could not determine active world: worlds directory %q is missing or unreadable: %w", worldsDir, err)
	}
	for _, e := range entries {
		if e.IsDir() {
			// return first match
			return filepath.Join(worldsDir, e.Name()), nil
		}
	}

	return "", fmt.Errorf("could not determine active world: no world directories found under %q", worldsDir)
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
	if err := scanner.Err(); err != nil {
		return "", err
	}

	// not found
	return "", nil
}
