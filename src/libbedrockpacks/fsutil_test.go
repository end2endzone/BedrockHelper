package libbedrockpacks

import (
	"os"
	"path/filepath"
	"testing"
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
		if got != want {
			t.Errorf("sanitizeCharactersInPath(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestMoveDir(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	err := os.MkdirAll(filepath.Join(src, "nested"), 0o755)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(src, "file.txt"), []byte("hello"), 0o644)
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(filepath.Join(src, "nested", "inner.txt"), []byte("world"), 0o644)
	if err != nil {
		t.Fatal(err)
	}

	dst := filepath.Join(t.TempDir(), "dst", "moved")
	err = moveDir(src, dst)
	if err != nil {
		t.Fatalf("moveDir failed: %v", err)
	}

	_, err = os.Stat(src)
	if !os.IsNotExist(err) {
		t.Errorf("expected source directory to be gone after move, stat err = %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dst, "file.txt"))
	if err != nil || string(data) != "hello" {
		t.Errorf("file.txt = %q, %v; want %q, nil", data, err, "hello")
	}
	data, err = os.ReadFile(filepath.Join(dst, "nested", "inner.txt"))
	if err != nil || string(data) != "world" {
		t.Errorf("nested/inner.txt = %q, %v; want %q, nil", data, err, "world")
	}
}
