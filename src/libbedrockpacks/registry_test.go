package libbedrockpacks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRegistryRoundTrip(t *testing.T) {
	worldDir := t.TempDir()

	// Reading a registry that doesn't exist yet returns an empty list, not an error.
	entries, err := readRegistry(worldDir, BehaviorPack)
	if err != nil {
		t.Fatalf("unexpected error reading missing registry: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("expected empty registry, got %v", entries)
	}

	uuid1 := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	uuid2 := "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"

	err = registerPack(worldDir, BehaviorPack, uuid1, Version{1, 0, 0})
	if err != nil {
		t.Fatalf("registerPack failed: %v", err)
	}
	err = registerPack(worldDir, BehaviorPack, uuid2, Version{2, 0, 0})
	if err != nil {
		t.Fatalf("registerPack failed: %v", err)
	}

	entries, err = readRegistry(worldDir, BehaviorPack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 registered packs, got %d: %v", len(entries), entries)
	}

	// Registering the same UUID again should update its version, not duplicate the entry.
	err = registerPack(worldDir, BehaviorPack, uuid1, Version{1, 5, 0})
	if err != nil {
		t.Fatalf("registerPack (update) failed: %v", err)
	}
	entries, _ = readRegistry(worldDir, BehaviorPack)
	if len(entries) != 2 {
		t.Fatalf("expected re-registering to update in place, got %d entries", len(entries))
	}
	found := false
	for _, e := range entries {
		if e.PackID == uuid1 {
			found = true
			if e.Version != (Version{1, 5, 0}) {
				t.Errorf("expected updated version [1 5 0], got %v", e.Version)
			}
		}
	}
	if !found {
		t.Fatal("expected uuid1 to still be present after update")
	}

	// Resource pack registry is independent from behavior pack registry.
	resourceEntries, err := readRegistry(worldDir, ResourcePack)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resourceEntries) != 0 {
		t.Fatalf("expected empty resource pack registry, got %v", resourceEntries)
	}

	err = unregisterPack(worldDir, BehaviorPack, uuid1)
	if err != nil {
		t.Fatalf("unregisterPack failed: %v", err)
	}
	entries, _ = readRegistry(worldDir, BehaviorPack)
	if len(entries) != 1 || entries[0].PackID != uuid2 {
		t.Fatalf("expected only uuid2 to remain, got %v", entries)
	}

	err = unregisterPack(worldDir, BehaviorPack, uuid1)
	if err == nil {
		t.Fatal("expected error unregistering an already-removed uuid, got nil")
	}
}

func TestRegistryFileNames(t *testing.T) {
	worldDir := t.TempDir()
	err := registerPack(worldDir, BehaviorPack, "uuid-b", Version{1, 0, 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = registerPack(worldDir, ResourcePack, "uuid-r", Version{1, 0, 0})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !fileExists(filepath.Join(worldDir, "world_behavior_packs.json")) {
		t.Error("expected world_behavior_packs.json to exist")
	}
	if !fileExists(filepath.Join(worldDir, "world_resource_packs.json")) {
		t.Error("expected world_resource_packs.json to exist")
	}
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
