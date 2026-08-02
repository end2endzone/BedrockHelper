package libbedrockpacks

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSanitizePackDirName(t *testing.T) {
	cases := map[string]string{
		"Foobar RP":            "Foobar RP",
		"":                     "pack",
		"   ":                  "pack",
		"Weird/Name:Here":      "Weird_Name_Here",
		"  Trim Me  ":          "Trim Me",
		"Path\\With*Bad?Chars": "Path_With_Bad_Chars",
	}
	for in, want := range cases {
		got := sanitizeCharactersInPath(in)
		require.Equal(t, want, got)
	}
}

func TestMoveDir(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	err := os.MkdirAll(filepath.Join(src, "nested"), 0o755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0o644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(src, "nested", "inner.txt"), []byte("world"), 0o644)
	require.NoError(t, err)

	dst := filepath.Join(t.TempDir(), "dst", "moved")
	err = moveDir(src, dst)
	require.NoError(t, err)

	_, err = os.Stat(src)
	require.True(t, os.IsNotExist(err), "expected source directory to be gone after move")

	data, err := os.ReadFile(filepath.Join(dst, "file.txt"))
	require.NoError(t, err)
	require.Equal(t, "hello", string(data))

	data, err = os.ReadFile(filepath.Join(dst, "nested", "inner.txt"))
	require.NoError(t, err)
	require.Equal(t, "world", string(data))
}
