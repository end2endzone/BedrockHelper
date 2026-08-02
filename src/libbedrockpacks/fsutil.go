package libbedrockpacks

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// moveDir moves a directory from `src` to `dst`.
// It first attempts an os.Rename which is the fastest way when `src` and `dst` are on the same filesystem.
// It falls back to a recursive copy-then-remove when that fails.
func moveDir(src string, dst string) error {
	// Try to rename first
	err := os.Rename(src, dst)
	if err == nil {
		return nil
	}

	// Fall back to a recursive copy-then-remove
	err = os.MkdirAll(filepath.Dir(dst), 0o755)
	if err != nil {
		return err
	}

	err = copyDir(src, dst)
	if err != nil {
		return err
	}

	return os.RemoveAll(src)
}

func copyDir(src, dst string) error {
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

		err = os.MkdirAll(filepath.Dir(target), 0o755)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		defer in.Close()

		out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode()|0o600)
		if err != nil {
			return err
		}
		defer out.Close()

		_, err = io.Copy(out, in)
		return err
	})
}

// sanitizeCharactersInPath converts the given string into a filesystem-safe directory name or file name.
// Characters that are not supported by filesystems are replaeced by an underscore.
func sanitizeCharactersInPath(name string) string {
	if strings.TrimSpace(name) == "" {
		return "pack"
	}
	replacer := strings.NewReplacer(
		"/", "_", "\\", "_", ":", "_", "*", "_", "?", "_",
		"\"", "_", "<", "_", ">", "_", "|", "_",
	)
	return strings.TrimSpace(replacer.Replace(name))
}

// copyFile copies a a file from `src` to `dst`.
// It first attempts an os.Rename which is the fastest way when `src` and `dst` are on the same filesystem.
// It falls back to a recursive copy-then-remove when that fails.
func copyFile(src string, dst string) error {
	dstDir := filepath.Dir(dst)

	// Create the directory tree (does nothing if it already exists)
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Open the source file
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// Create the destination file
	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy the contents
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	return nil
}

// fileExists checks if a file exists for the given path
func fileExists(path string) bool {
	_ /*info*/, err := os.Stat(path)
	return err == nil
}
