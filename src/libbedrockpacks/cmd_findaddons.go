package libminecraftbedrock

import (
	"os"
	"path/filepath"
)

// FindAddonsInDir searches dir (resursively, if recursive is true) for files that are valid add-on packs.
// It returns their absolute paths.
func FindAddonsInDir(dir string, recursive bool) ([]string, error) {
	var results []string

	if recursive {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				// Skip unreadable entries rather than aborting the whole scan.
				return nil
			}
			if info.IsDir() {
				// skip directories
				return nil
			}

			// check file extension
			if !IsValidAddonFileExtension(path) {
				// skip invalid file extension
				return nil
			}

			// run full check for file
			if IsValidAddonFile(path) {
				abs, err := filepath.Abs(path)
				if err != nil {
					abs = path
				}
				results = append(results, abs)
			}
			return nil
		})

		if err != nil {
			return nil, err
		}

		return results, nil
	}

	// non-recursive lookup
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if e.IsDir() {
			// skip directories
			continue
		}

		// check file extension
		path := filepath.Join(dir, e.Name())
		if !IsValidAddonFileExtension(path) {
			// skip invalid file extension
			continue
		}

		// run full check for file
		if IsValidAddonFile(path) {
			abs, err := filepath.Abs(path)
			if err != nil {
				abs = path
			}

			// valid result
			results = append(results, abs)
		}
	}

	return results, nil
}
