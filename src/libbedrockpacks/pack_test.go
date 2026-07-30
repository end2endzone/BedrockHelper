package libbedrockpacks

import (
	"os"
	"testing"
)

func TestListInstalledPacks(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")
	defer os.RemoveAll(tempServerDir)

	packs, err := ListInstalledPacks(tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packs) != 1 {
		t.Fatalf("expected 1 registered pack, got %d: %+v", len(packs), packs)
	}
	if packs[0].UUID != "2bda6085-9d71-4d8a-9b9f-74e07b30459c" {
		t.Errorf("UUID = %q, want the registered uuid", packs[0].UUID)
	}
	if packs[0].Name != "Foobar BP" {
		t.Errorf("Name = %q, want %q (resolved from installed manifest.json)", packs[0].Name, "Foobar BP")
	}
	if packs[0].Kind != BehaviorPack {
		t.Errorf("Kind = %v, want BehaviorPack", packs[0].Kind)
	}
}

func TestListInstalledPacks_EmptyServer(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	packs, err := ListInstalledPacks(tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packs) != 0 {
		t.Errorf("expected no registered packs, got %+v", packs)
	}
}

func TestListInstalledPacks_AfterInstallAndUninstall(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	packs, err := ListInstalledPacks(tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packs) != 2 {
		t.Fatalf("expected 2 registered packs after install, got %d", len(packs))
	}

	_, err = UninstallAddon(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}
	packs, err = ListInstalledPacks(tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(packs) != 0 {
		t.Errorf("expected 0 registered packs after uninstall, got %d: %+v", len(packs), packs)
	}
}

func TestListInstalledPacks_InvalidServer(t *testing.T) {
	invalidServer := getServerFixturePath(t, "not_a_server")
	_, err := ListInstalledPacks(invalidServer)
	if err == nil {
		t.Fatal("expected error for invalid server directory, got nil")
	}
}
