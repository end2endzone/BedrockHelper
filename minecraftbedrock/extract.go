package minecraftbedrock

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

// FindManifestsRelativePathInAddon returns the archive-relative paths of every manifest.json file found inside the given archive add-on file.
// Note that returned paths are using forward slashes, as stored in the zip archive.
// Returns an error if no manifest.json is found in the addon.
func FindManifestsRelativePathInAddon(addonPath string) ([]string, error) {
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
		return nil, fmt.Errorf("no manifest.json file found inside %q", addonPath)
	}

	return manifests, nil
}

// readZipEntry reads and returns the raw bytes of a single named entry inside a zip archive.
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

// ZipFilePathJoin joins any number of path elements into a single path, separating them with a the ZIP specific separator `/`.
// Empty elements are ignored.
func ZipFilePathJoin(elem ...string) string {
	var b strings.Builder
	for _, value := range elem {
		// Empty elements are ignored
		if value == "" {
			continue
		}

		// Zip file does not supports `.` and `..` directories
		if value == "." {
			continue
		}

		if b.Len() == 0 {
			// first elements is the result
			b.WriteString(value)
		} else {
			// following elements must have a separator
			b.WriteString("/")
			b.WriteString(value)
		}
	}
	result := b.String()
	return result
}

// ZipFilePathGetParentDir returns the parent directory of a given path.
// The function supports both forward slashes and backslashes.
func ZipFilePathGetParentDir(path string) string {
	// Clean up trailing slashes so we don't look at the same folder
	for len(path) > 1 && (path[len(path)-1] == '/' || path[len(path)-1] == '\\') {
		path = path[:len(path)-1]
	}

	// Find the last forward slash or backslash
	lastSlash := strings.LastIndexAny(path, "/\\")

	// If no slash is found, there is no parent directory
	if lastSlash == -1 {
		return "." // return "." to have the same behavior as filepath.Dir()
	}

	// If the last slash is at the very beginning, return the root slash
	if lastSlash == 0 {
		return string(path[0])
	}

	// Return everything up to the last slash, exclusing the last slash
	return path[:lastSlash]
}
