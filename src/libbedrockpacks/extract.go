package libbedrockpacks

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// ExtractZip extracts the contents of the given file path into destDir.
// The destDir is created if it does not exist.
func ExtractZip(path string, destDir string) error {
	r, err := zip.OpenReader(path)
	if err != nil {
		return fmt.Errorf("failed to open add-on %q: %w", path, err)
	}
	defer r.Close()

	destDir = filepath.Clean(destDir)

	// Create destination directory if not exists
	err = os.MkdirAll(destDir, 0o755)
	if err != nil {
		return fmt.Errorf("failed to create extraction directory %q: %w", destDir, err)
	}

	// For each file in the archive
	for _, f := range r.File {
		targetPath := filepath.Join(destDir, filepath.FromSlash(f.Name))

		// Zip-slip protection: reject entries with invalid names that would escape destDir.
		if !strings.HasPrefix(targetPath, destDir+string(os.PathSeparator)) && targetPath != destDir {
			return fmt.Errorf("add-on %q contains an unsafe path %q", path, f.Name)
		}

		if f.FileInfo().IsDir() {
			// Create destination directory if not exists
			err := os.MkdirAll(targetPath, 0o755)
			if err != nil {
				return err
			}
			continue
		}

		// Create destination directory if not exists
		err := os.MkdirAll(filepath.Dir(targetPath), 0o755)
		if err != nil {
			return err
		}

		// Call utility function to process a save a single file
		err = extractOneFile(f, targetPath)
		if err != nil {
			return fmt.Errorf("failed to extract %q from %q: %w", f.Name, path, err)
		}
	}

	return nil
}

func extractOneFile(f *zip.File, targetPath string) error {
	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	out, err := os.OpenFile(targetPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode()|0o600)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, rc)
	if err != nil {
		return err
	}
	return nil
}

// FindManifestsInAddon returns the archive-relative paths of every manifest.json file found inside the given archive add-on file.
// Note that returned paths are using forward slashes, as stored in the zip archive.
// Returns an error if no manifest.json is found in the addon.
func FindManifestsInAddon(addonPath string) ([]string, error) {
	r, err := zip.OpenReader(addonPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open add-on %q: %w", addonPath, err)
	}
	defer r.Close()

	var manifests []string
	for _, f := range r.File {
		if f.FileInfo().IsDir() {
			// Not interrested in directories
			continue
		}
		if filepath.Base(f.Name) == "manifest.json" {
			// Found a manifest.json file
			manifests = append(manifests, f.Name)
		}
	}

	if len(manifests) == 0 {
		return nil, fmt.Errorf("no manifest.json files found inside %q", addonPath)
	}

	return manifests, nil
}

// readZipEntry reads and returns the raw bytes of a single named entry
// inside a zip archive.
func readZipEntry(zipPath, entryName string) ([]byte, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	for _, f := range r.File {
		if f.Name == entryName {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}

	return nil, fmt.Errorf("entry %q not found in %q", entryName, zipPath)
}

// filterManifestPaths filters the given list of relative manifest.json paths to only keep the manifest that
// matches a behavior pack or a resource pack.
// Excludes manifests that are located at the root of an .mcpaddon file.
func filterManifestPaths(manifestPaths []string) []string {
	// Nothing to filter if there is only 1 manifest found
	if len(manifestPaths) <= 1 {
		return manifestPaths
	}

	hasNested := false
	for _, p := range manifestPaths {
		if strings.Contains(p, "/") {
			hasNested = true
			break
		}
	}
	if !hasNested {
		return manifestPaths
	}

	var filtered []string
	for _, p := range manifestPaths {
		if !strings.Contains(p, "/") {
			// Root manifest alongside nested pack manifests, skip it.
			continue
		}
		filtered = append(filtered, p)
	}

	return filtered
}
