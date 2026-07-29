package libbedrockpacks

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

// testdataDir returns the absolute path to the testdata directory located at the project's root directory.
// It returns the directory location regardless of which package's test binary is running.
func testdataDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}

	// This file lives in <module>/libbedrockpacks, so testdata is a sibling
	// of that directory.
	return filepath.Join(wd, "..", "testdata")
}

func addonPath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testdataDir(t), "addons", name)
}

func serverFixturePath(t *testing.T, name string) string {
	t.Helper()
	return filepath.Join(testdataDir(t), "servers", name)
}

// copyServerFixture copies the given testdata server directory into a temporary directory.
// This is required to prevent calling function that would modify the directory. It makes sure we never affect the testdata files under CM.
// Returns the path to the temporary copy. The caller is reponsible to delete the returned temporary directory.
func copyServerFixture(t *testing.T, name string) string {
	t.Helper()
	src := serverFixturePath(t, name)
	dst := filepath.Join(t.TempDir(), name)
	if err := copyTree(src, dst); err != nil {
		t.Fatalf("failed to copy server fixture %q: %v", name, err)
	}
	return dst
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()
		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
		if err != nil {
			return err
		}
		defer out.Close()
		_, err = io.Copy(out, in)
		return err
	})
}
