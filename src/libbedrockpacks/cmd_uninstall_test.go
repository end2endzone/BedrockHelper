package libbedrockpacks

import (
	"os"
	"testing"
)

func TestUninstallAddon(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	_, err := InstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("setup install failed: %v", err)
	}

	uninstalledPacks, err := UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(uninstalledPacks) != 2 {
		t.Fatalf("expected 2 packs uninstalled, got %d", len(uninstalledPacks))
	}

	// Verify both registry files were deleted.
	worldDir, _ := FindActiveWorldDir(tempServerDir)
	bpEntries, _ := readRegistry(worldDir, BehaviorPack)
	rpEntries, _ := readRegistry(worldDir, ResourcePack)
	if len(bpEntries) != 0 || len(rpEntries) != 0 {
		t.Errorf("expected empty registries after uninstall, got bp=%v rp=%v", bpEntries, rpEntries)
	}

	// Assert that each uninstalled pack's directory were deleted
	for _, p := range uninstalledPacks {
		_, err := os.Stat(p.Path)
		if !os.IsNotExist(err) {
			t.Errorf("expected pack directory %q to be removed, stat err = %v", p.Path, err)
		}
	}
}

func TestUninstallAddon_NotInstalled(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	// Never installed, so this should fail to find the pack in the world.
	_, err := UninstallAddonInServer(getAddonFixturePath(t, "foobar.mcaddon"), tempServerDir)
	if err == nil {
		t.Fatal("expected error uninstalling a pack that was never installed, got nil")
	}
}

func TestUninstallPackByUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")
	defer os.RemoveAll(tempServerDir)

	pack, err := UninstallPackInServerByUUID("2bda6085-9d71-4d8a-9b9f-74e07b30459c", tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pack.Name() != "Foobar BP" {
		t.Errorf("Name = %q, want %q", pack.Name(), "Foobar BP")
	}
	if pack.KindSafe() != BehaviorPack {
		t.Errorf("Kind = %v, want BehaviorPack", pack.KindSafe())
	}

	worldDir, _ := FindActiveWorldDir(tempServerDir)
	entries, _ := readRegistry(worldDir, BehaviorPack)
	if len(entries) != 0 {
		t.Errorf("expected registry to be empty after uninstall, got %v", entries)
	}
	_, err = os.Stat(pack.Path)
	if !os.IsNotExist(err) {
		t.Errorf("expected pack directory to be removed")
	}
}

func TestUninstallPackByUUID_UnknownUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_with_installed_pack")
	defer os.RemoveAll(tempServerDir)

	_, err := UninstallPackInServerByUUID("00000000-0000-0000-0000-000000000000", tempServerDir)
	if err == nil {
		t.Fatal("expected error for unknown uuid, got nil")
	}
}

func TestInstallThenUninstallByUUID(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	installed, err := InstallAddonInServer(getAddonFixturePath(t, "behavior_only.mcpack"), tempServerDir)
	if err != nil {
		t.Fatalf("install failed: %v", err)
	}
	if len(installed) != 1 {
		t.Fatalf("expected 1 installed pack, got %d", len(installed))
	}

	pack, err := UninstallPackInServerByUUID(installed[0].UUID(), tempServerDir)
	if err != nil {
		t.Fatalf("uninstall by uuid failed: %v", err)
	}
	if pack.Name() != "Solo BP" {
		t.Errorf("Name = %q, want %q", pack.Name(), "Solo BP")
	}
}
