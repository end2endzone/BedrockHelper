package minecraftbedrock

import (
	"os"
	"path/filepath"
)

func processFileAddonEntry(path string, results *[]string) {
	// Check file extension
	if !IsValidAddonFileExtension(path) {
		// Invalid
		return
	}

	// Run full check for file
	if !IsValidAddonFile(path) {
		// Invalid
		return
	}

	// Valid. Add to matching results
	*results = append(*results, path)
}

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

			// Make the path absolute
			abs, err := filepath.Abs(path)
			if err == nil {
				path = abs
			}

			// Process entry
			processFileAddonEntry(path, &results)

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

		// Make the path absolute
		path := filepath.Join(dir, e.Name())

		// Process entry
		processFileAddonEntry(path, &results)
	}

	return results, nil
}
