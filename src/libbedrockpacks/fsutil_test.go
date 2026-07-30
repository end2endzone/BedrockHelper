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
		got := sanitizePackDirName(in)
		if got != want {
			t.Errorf("sanitizePackDirName(%q) = %q, want %q", in, got, want)
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

func TestFindPackDirByUUID(t *testing.T) {
	newServerDir := copyServerFixture(t, "server_with_installed_pack")
	worldDir, err := FindActiveWorldDir(newServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dir, kind, err := findPackInstallDirByUUID(worldDir, "2bda6085-9d71-4d8a-9b9f-74e07b30459c")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if kind != BehaviorPack {
		t.Errorf("kind = %v, want BehaviorPack", kind)
	}
	if filepath.Base(dir) != "Foobar BP" {
		t.Errorf("dir = %q, want basename %q", dir, "Foobar BP")
	}

	_, _, err = findPackInstallDirByUUID(worldDir, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for unknown uuid, got nil")
	}
}
