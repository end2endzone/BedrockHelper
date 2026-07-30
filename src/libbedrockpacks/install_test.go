package libbedrockpacks

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func TestInstallAddon_Bundle(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	installedPacks, err := InstallAddon(addonPath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(installedPacks) != 2 {
		t.Fatalf("expected 2 packs installedPacks, got %d: %+v", len(installedPacks), installedPacks)
	}

	var kinds []string
	for _, pack := range installedPacks {
		kinds = append(kinds, pack.Kind.String())
		_, err := os.Stat(filepath.Join(pack.Directory, "manifest.json"))
		if err != nil {
			t.Errorf("expected manifest.json at %q: %v", pack.Directory, err)
		}
	}

	sort.Strings(kinds)
	if kinds[0] != "BehaviorPack" || kinds[1] != "ResourcePack" {
		t.Errorf("kinds = %v, want [BehaviorPack ResourcePack]", kinds)
	}

	// Verify both registry files were updated.
	worldDir, _ := FindActiveWorldDir(tempServerDir)
	bpEntries, err := readRegistry(worldDir, BehaviorPack)
	if err != nil || len(bpEntries) != 1 {
		t.Errorf("behavior pack registry = %v, err %v; want 1 entry", bpEntries, err)
	}
	rpEntries, err := readRegistry(worldDir, ResourcePack)
	if err != nil || len(rpEntries) != 1 {
		t.Errorf("resource pack registry = %v, err %v; want 1 entry", rpEntries, err)
	}

	// The master link manifest.json must NOT itself be treated as an installed pack.
	for _, p := range installedPacks {
		if p.Name == "Foobar Addon" {
			t.Errorf("master link manifest should not be installed as a pack: %v", p)
		}
	}
}

func TestInstallAddon_StandaloneMcpack(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server_no_level_name")
	defer os.RemoveAll(tempServerDir)

	installedPacks, err := InstallAddon(addonPath(t, "solo.mcpack"), tempServerDir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(installedPacks) != 1 {
		t.Fatalf("expected 1 pack installedPacks, got %d", len(installedPacks))
	}

	if installedPacks[0].Kind != ResourcePack {
		t.Errorf("Kind = %v, want ResourcePack", installedPacks[0].Kind)
	}
	if installedPacks[0].Name != "Solo RP" {
		t.Errorf("Name = %q, want %q", installedPacks[0].Name, "Solo RP")
	}
}

func TestInstallAddon_ReinstallReplacesExisting(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	_, err := InstallAddon(addonPath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("first install failed: %v", err)
	}

	installedPacks, err := InstallAddon(addonPath(t, "foobar.mcaddon"), tempServerDir)
	if err != nil {
		t.Fatalf("reinstall failed: %v", err)
	}
	if len(installedPacks) != 2 {
		t.Fatalf("expected 2 packs after reinstall, got %d", len(installedPacks))
	}

	worldDir, _ := FindActiveWorldDir(tempServerDir)
	bpEntries, _ := readRegistry(worldDir, BehaviorPack)
	if len(bpEntries) != 1 {
		t.Errorf("expected reinstall to not duplicate registry entries, got %d", len(bpEntries))
	}
}

func TestInstallAddon_Errors(t *testing.T) {
	tempServerDir := copyServerFixture(t, "server")
	defer os.RemoveAll(tempServerDir)

	t.Run("invalid addon file", func(t *testing.T) {
		_, err := InstallAddon(addonPath(t, "corrupt.mcaddon"), tempServerDir)
		if err == nil {
			t.Fatal("expected error installing a corrupt add-on, got nil")
		}
	})

	t.Run("addon with no manifests", func(t *testing.T) {
		_, err := InstallAddon(addonPath(t, "no_manifest.zip"), tempServerDir)
		if err == nil {
			t.Fatal("expected error installing an add-on with no manifests, got nil")
		}
	})

	t.Run("invalid server directory", func(t *testing.T) {
		newInvalidServerDir := serverFixturePath(t, "not_a_server")
		_, err := InstallAddon(addonPath(t, "foobar.mcaddon"), newInvalidServerDir)
		if err == nil {
			t.Fatal("expected error installing to an invalid server directory, got nil")
		}
	})

	t.Run("server with level-name set but worlds dir not yet created should fail", func(t *testing.T) {
		// InstallAddon should create directory `worlds/<level-name>/behavior_packs` or `worlds/<level-name>/resource_packs`
		// on the fly even if they don't exist yet.
		tempServerDir := copyServerFixture(t, "not_a_server_missing_worlds")
		defer os.RemoveAll(tempServerDir)

		_ /*installedPacks*/, err := InstallAddon(addonPath(t, "foobar.mcaddon"), tempServerDir)
		if err == nil {
			t.Fatalf("expected an error, got nil")
		}
	})
}
