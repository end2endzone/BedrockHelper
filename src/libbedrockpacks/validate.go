package libbedrockpacks

import (
	"fmt"
	"os"
	"path/filepath"
)

// validAddonExtensions lists the file extensions accepted by the library.
var validAddonExtensions = map[string]bool{
	".zip":     true,
	".mcaddon": true,
	".mcpack":  true,
}

// IsServerDirectory validates that the given path is a directory that matches a Minecraft Bedrock dedicated server installation.
// It is considered valid if it contains the following:
// * a server.properties file
// * a bedrock_server executable
// * a worlds directory.
func IsServerDirectory(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("server location %q does not exist: %w", path, err)
	}
	if !info.IsDir() {
		return false, fmt.Errorf("server location %q is not a directory", path)
	}

	// Check mandatory files/directories
	candidates := []string{"server.properties", "worlds"}
	for _, c := range candidates {
		_, err := os.Stat(filepath.Join(path, c))
		if err != nil {
			// assume file/dir is missing
			// path is not a server install directory
			return false, nil
		}
	}

	// Check optional (or) candidates
	candidates = []string{"bedrock_server", "bedrock_server.exe"}
	execFound := false
	for _, c := range candidates {
		_, err := os.Stat(filepath.Join(path, c))
		if err != nil {
			execFound = true
		}
	}
	if !execFound {
		// exec file is missing
		// path is not a server install directory
		return false, nil
	}

	return true, nil
}
