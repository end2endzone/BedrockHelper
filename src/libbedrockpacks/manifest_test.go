package libbedrockpacks

import "testing"

func TestParseManifest(t *testing.T) {
	valid := []byte(`{
		"format_version": 2,
		"header": {"name": "Test Pack", "uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "version": [1,2,3]},
		"modules": [{"type": "data", "uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "version": [1,2,3]}]
	}`)

	m, err := LoadManifestFromBytes(valid)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.Header.Name != "Test Pack" {
		t.Errorf("Name = %q, want %q", m.Header.Name, "Test Pack")
	}
	if m.Header.UUID != "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa" {
		t.Errorf("UUID = %q, want the header uuid", m.Header.UUID)
	}
	if m.Header.Version != (Version{1, 2, 3}) {
		t.Errorf("Version = %v, want [1 2 3]", m.Header.Version)
	}
	got := m.Header.Version.String()
	want := "1.2.3"
	if got != want {
		t.Errorf("Version.String() = %q, want %q", got, want)
	}

	t.Run("missing uuid", func(t *testing.T) {
		data := []byte(`{"header": {"name": "No UUID", "version": [1,0,0]}}`)
		_, err := LoadManifestFromBytes(data)
		if err == nil {
			t.Fatal("expected error for manifest missing header.uuid, got nil")
		}
	})

	t.Run("malformed json", func(t *testing.T) {
		_, err := LoadManifestFromBytes([]byte(`{not valid json`))
		if err == nil {
			t.Fatal("expected error for malformed JSON, got nil")
		}
	})
}

func TestIdentifyPackKind(t *testing.T) {
	cases := []struct {
		name       string
		moduleType string
		want       PackKind
	}{
		{"data module -> behavior pack", "data", BehaviorPack},
		{"script module -> behavior pack", "script", BehaviorPack},
		{"resources module -> resource pack", "resources", ResourcePack},
		{"client_data module -> resource pack", "client_data", ResourcePack},
		{"interface module -> resource pack", "interface", ResourcePack},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := &AddonManifest{
				Header:  Header{UUID: "uuid", Name: "n"},
				Modules: []Module{{Type: tc.moduleType, UUID: "mod-uuid"}},
			}
			kind, err := IdentifyPackKind(m)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if kind != tc.want {
				t.Errorf("IdentifyPackKind() = %v, want %v", kind, tc.want)
			}
		})
	}

	t.Run("no modules", func(t *testing.T) {
		m := &AddonManifest{Header: Header{UUID: "uuid"}}
		_, err := IdentifyPackKind(m)
		if err == nil {
			t.Fatal("expected error for manifest with no modules, got nil")
		}
	})

	t.Run("unrecognized module type", func(t *testing.T) {
		m := &AddonManifest{
			Header:  Header{UUID: "uuid"},
			Modules: []Module{{Type: "something_else"}},
		}
		_, err := IdentifyPackKind(m)
		if err == nil {
			t.Fatal("expected error for unrecognized module type, got nil")
		}
	})

	t.Run("nil manifest", func(t *testing.T) {
		_, err := IdentifyPackKind(nil)
		if err == nil {
			t.Fatal("expected error for nil manifest, got nil")
		}
	})

	t.Run("mixed modules prefer behavior", func(t *testing.T) {
		m := &AddonManifest{
			Header: Header{UUID: "uuid"},
			Modules: []Module{
				{Type: "resources"},
				{Type: "data"},
			},
		}
		kind, err := IdentifyPackKind(m)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if kind != BehaviorPack {
			t.Errorf("IdentifyPackKind() with mixed modules = %v, want BehaviorPack", kind)
		}
	})
}

func TestPackKindString(t *testing.T) {
	cases := map[PackKind]string{
		BehaviorPack: "BehaviorPack",
		ResourcePack: "ResourcePack",
		UnknownPack:  "UnknownPack",
	}
	for kind, want := range cases {
		got := kind.String()
		if got != want {
			t.Errorf("%v.String() = %q, want %q", int(kind), got, want)
		}
	}
}
