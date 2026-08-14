package libminecraftbedrock

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

// IsValidAddonFileExtension checks if the given file path has a valid file extension for an add-on.
func IsValidAddonFileExtension(path string) bool {
	// check file extension
	if !validAddonExtensions[strings.ToLower(filepath.Ext(path))] {
		// skip invalid file extension
		return false
	}
	return true
}

// ValidateServerDirectory asserts that the given path is a directory that matches a Minecraft Bedrock dedicated server installation.
// It is considered valid if it contains the following:
// * a server.properties file
// * a bedrock_server executable
// * a worlds directory.
// Returns nil if the given path is server. Returns a valid error otherwise.
func ValidateServerDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("server path %q does not exist: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("server path is not a directory: %q", path)
	}

	// Check mandatory files/directories
	candidates := []string{"server.properties", "worlds"}
	for _, c := range candidates {
		_, err := os.Stat(filepath.Join(path, c))
		if err != nil {
			return fmt.Errorf("server path %q is missing files: %w", path, err)
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
		return fmt.Errorf("server path %q is missing the executable file", path)
	}

	return nil
}

// IsValidServerDirectory validates that the given path is a directory that matches a Minecraft Bedrock dedicated server installation.
// It is considered valid if it contains the following:
// * a server.properties file
// * a bedrock_server executable
// * a worlds directory.
// Returns true if the given path is valid server. Returns false otherwise.
func IsValidServerDirectory(path string) bool {
	err := ValidateServerDirectory(path)
	if err != nil {
		// something is wront, not a valid server directory
		return false
	}
	return true
}

// ValidateAddonFile asserts that the given path is a valid addon file.
// Addon files are zip file archives.
// The file extension must be one of the following:
// `.zip`, `.mcaddon` or `.mcpack`.
func ValidateAddonFile(path string) error {
	// Check the file's path
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("add-on file %q does not exist: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("add-on file path is a directory: %q", path)
	}

	// Check file extension
	if !IsValidAddonFileExtension(path) {
		ext := filepath.Ext(path)
		return fmt.Errorf("add-on file %q has an unsupported file extension: %q", path, ext)
	}

	// Check zip archive
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("add-on file %q is not a valid zip archive: %w", path, err)
	}
	defer r.Close()

	return nil
}

// IsValidAddonFile validates that the given path is a directory that matches a Minecraft Bedrock dedicated server installation.
// Addon files are zip file archives.
// The file extension must be one of the following:
// `.zip`, `.mcaddon` or `.mcpack`.
// Returns true if the given path is a valid add-on. Returns false otherwise.
func IsValidAddonFile(path string) bool {
	err := ValidateAddonFile(path)
	if err != nil {
		// something is wront, not a valid add-on file
		return false
	}
	return true
}

// ValidateDirectory asserts that the given path is a valid directory that exists.
func ValidateDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("directory %q is not found: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %q", path)
	}

	return nil
}

// IsValidMcPackFile checks that a given file path is a valid MCPack file.
func IsValidMcPackFile(path string) bool {
	// Check by file extension
	valid := IsValidAddonFileExtension(path)
	if !valid {
		return false
	}

	// Check by peeking at the content
	valid = IsValidAddonFile(path)
	if !valid {
		return false
	}

	// Check with manifests
	manifestPaths, err := FindManifestsRelativePathInAddon(path)
	if err != nil {
		return false
	}
	if len(manifestPaths) != 1 {
		return false
	}

	// Check that manifests is at root directory
	parentDir := filepath.Dir(manifestPaths[0])
	if parentDir != "." {
		// Not a MCPack.
		// This manifest is not at the relative root directory (such as a MCAddon)
		return false
	}

	return true
}

// IsValidMcAddonFile checks that a given file path is a valid MCAddon file.
func IsValidMcAddonFile(path string) bool {
	// Check by file extension
	valid := IsValidAddonFileExtension(path)
	if !valid {
		return false
	}

	// Check by peeking at the content
	valid = IsValidAddonFile(path)
	if !valid {
		return false
	}

	// Check with manifests
	manifestPaths, err := FindManifestsRelativePathInAddon(path)
	if err != nil {
		return false
	}
	if len(manifestPaths) == 0 {
		return false
	}

	// Check that each manifests's file is at root directory
	for _, manifestPath := range manifestPaths {
		parentDir := filepath.Dir(manifestPath)
		if parentDir == "." {
			// Not a MCAddon.
			// This manifest is at the relative root directory (such as a MCPack)
			return false
		}
	}

	return true
}
