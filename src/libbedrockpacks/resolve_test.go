package libbedrockpacks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolvePackByUUID(t *testing.T) {
	// This test copy an add-on file inside the server directory so that we can resolve from a UUID.

	// Create a tempoerary copy of the 'server' testdata server.
	// Defer removal of the temporary directory when the function returns
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	// Copy an add-on file inside the server directory so that we can resolve from a UUID.
	src, err := os.ReadFile(getAddonFixturePath(t, "foobar.mcaddon"))
	if err != nil {
		t.Fatalf("failed to read fixture addon: %v", err)
	}
	newServerAddonsSubDir := filepath.Join(tempServerDir, "new_incoming")
	if err := os.MkdirAll(newServerAddonsSubDir, 0o755); err != nil {
		t.Fatal(err)
	}
	newAddonPath := filepath.Join(newServerAddonsSubDir, "foobar.mcaddon")
	if err := os.WriteFile(newAddonPath, src, 0o644); err != nil {
		t.Fatal(err)
	}

	foobarMcAddonRessourcePackUUID := "33333333-3333-3333-3333-333333333333"

	// Try to resolve
	got, err := ResolvePackByUUID(foobarMcAddonRessourcePackUUID, tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Assert we found our pack inside our addon
	if got != newAddonPath {
		t.Errorf("ResolvePackByUUID() = %q, want %q", got, newAddonPath)
	}
}

func TestResolvePackByUUID_NotFound(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	_, err := ResolvePackByUUID("00000000-0000-0000-0000-000000000000", tempServerDir)
	if err == nil {
		t.Fatal("expected error when no add-on matches the uuid, got nil")
	}
}

func TestResolvePackByUUID_InvalidServer(t *testing.T) {
	tempServerDir := copyServerFixture(t, "not_a_server")
	defer os.RemoveAll(tempServerDir)

	path, err := ResolvePackByUUID("33333333-3333-3333-3333-333333333333", tempServerDir)
	if err == nil {
		t.Fatalf("expected an error, got nil")
	}
	if path != "" {
		t.Fatalf("expected an empty path, got: %v", path)
	}
}
