package libbedrockpacks

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestFindManifestsInAddon(t *testing.T) {
	t.Run("bundle with master + two packs", func(t *testing.T) {
		got, err := FindManifestsInAddon(getAddonFixturePath(t, "foobar.mcaddon"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sort.Strings(got)
		want := []string{"foobar_BP/manifest.json", "foobar_RP/manifest.json", "manifest.json"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindManifestsInAddon() = %v, want %v", got, want)
		}
	})

	t.Run("standalone mcpack", func(t *testing.T) {
		got, err := FindManifestsInAddon(getAddonFixturePath(t, "solo.mcpack"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sort.Strings(got)
		want := []string{"manifest.json"}

		if !reflect.DeepEqual(got, want) {
			t.Errorf("FindManifestsInAddon() = %v, want %v", got, want)
		}
	})

	t.Run("zip with no manifest", func(t *testing.T) {
		_, err := FindManifestsInAddon(getAddonFixturePath(t, "no_manifest.zip"))
		if err == nil {
			t.Fatal("expected error for zip with no manifest.json, got nil")
		}
	})

	t.Run("nonexistent file", func(t *testing.T) {
		_, err := FindManifestsInAddon("/tmp/nope.mcaddon")
		if err == nil {
			t.Fatal("expected error for nonexistent file, got nil")
		}
	})
}

func TestPackManifestPaths(t *testing.T) {
	cases := []struct {
		name  string
		input []string
		want  []string
	}{
		{
			name:  "bundle: master manifest filtered out",
			input: []string{"manifest.json", "foobar_BP/manifest.json", "foobar_RP/manifest.json"},
			want:  []string{"foobar_BP/manifest.json", "foobar_RP/manifest.json"},
		},
		{
			name:  "standalone pack: single root manifest kept",
			input: []string{"manifest.json"},
			want:  []string{"manifest.json"},
		},
		{
			name:  "no nesting at all: all kept as-is",
			input: []string{"manifest.json", "other/manifest2.json"},
			want:  []string{"other/manifest2.json"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterManifestPaths(tc.input)

			sort.Strings(got)
			want := append([]string(nil), tc.want...)

			sort.Strings(want)
			if !reflect.DeepEqual(got, want) {
				t.Errorf("filterManifestPaths(%v) = %v, want %v", tc.input, got, want)
			}
		})
	}
}

func TestExtractAddon(t *testing.T) {
	dest := t.TempDir()
	err := ExtractZip(getAddonFixturePath(t, "foobar.mcaddon"), dest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, rel := range []string{
		"manifest.json",
		"foobar_BP/manifest.json",
		"foobar_BP/items/coin.json",
		"foobar_RP/manifest.json",
		"foobar_RP/textures/items/coin.png",
	} {
		_, err := os.Stat(filepath.Join(dest, filepath.FromSlash(rel)))
		if err != nil {
			t.Errorf("expected extracted file %q to exist: %v", rel, err)
		}
	}
}

func TestReadZipEntry(t *testing.T) {
	data, err := readZipEntry(getAddonFixturePath(t, "solo.mcpack"), "manifest.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, err := LoadManifestFromBytes(data)
	if err != nil {
		t.Fatalf("failed to parse extracted manifest: %v", err)
	}

	if m.Header.Name != "Solo RP" {
		t.Errorf("Name = %q, want %q", m.Header.Name, "Solo RP")
	}

	_, err = readZipEntry(getAddonFixturePath(t, "solo.mcpack"), "does/not/exist.json")
	if err == nil {
		t.Fatal("expected error reading a nonexistent zip entry, got nil")
	}
}
