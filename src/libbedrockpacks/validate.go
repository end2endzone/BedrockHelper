package libbedrockpacks

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// validAddonExtensions lists the file extensions accepted by the library.
var validAddonExtensions = map[string]bool{
	".zip":     true,
	".mcaddon": true,
	".mcpack":  true,
}

// IsValidServerDirectory validates that the given path is a directory that matches a Minecraft Bedrock dedicated server installation.
// It is considered valid if it contains the following:
// * a server.properties file
// * a bedrock_server executable
// * a worlds directory.
func IsValidServerDirectory(path string) (bool, error) {
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
		if err == nil {
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

// IsValidAddonFile validates that the given path is a valid addon file.
// Addon files are zip files with one of the following accepted file extensions:
// `.zip`, `.mcaddon` or `.mcpack`.
func IsValidAddonFile(path string) (bool, error) {
	// Check the file's path
	info, err := os.Stat(path)
	if err != nil {
		return false, fmt.Errorf("add-on file %q does not exist: %w", path, err)
	}
	if info.IsDir() {
		return false, fmt.Errorf("%q is a directory, not an add-on file", path)
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(path))
	if !validAddonExtensions[ext] {
		return false, fmt.Errorf("%q does not have a supported add-on extension (.zip, .mcaddon, .mcpack)", path)
	}

	// Check zip archive
	r, err := zip.OpenReader(path)
	if err != nil {
		return false, fmt.Errorf("%q is not a valid zip archive: %w", path, err)
	}
	defer r.Close()

	return true, nil
}
