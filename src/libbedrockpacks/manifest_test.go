package libbedrockpacks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseManifest(t *testing.T) {
	valid := []byte(`{
		"format_version": 2,
		"header": {"name": "Test Pack", "uuid": "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", "version": [1,2,3]},
		"modules": [{"type": "data", "uuid": "bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb", "version": [1,2,3]}]
	}`)

	m, err := LoadManifestFromBytes(valid)
	require.NoError(t, err)
	require.Equal(t, "Test Pack", m.Header.Name)
	require.Equal(t, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa", m.Header.UUID)
	require.Equal(t, (Version{1, 2, 3}), m.Header.Version)
	require.Equal(t, "1.2.3", m.Header.Version.String())

	t.Run("missing uuid", func(t *testing.T) {
		data := []byte(`{"header": {"name": "No UUID", "version": [1,0,0]}}`)
		_, err := LoadManifestFromBytes(data)
		require.Error(t, err, "expected error for manifest missing header.uuid, got nil")
	})

	t.Run("malformed json", func(t *testing.T) {
		_, err := LoadManifestFromBytes([]byte(`{not valid json`))
		require.Error(t, err, "expected error for malformed JSON, got nil")
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
			require.NoError(t, err)
			require.Equal(t, tc.want, kind)
		})
	}

	t.Run("no modules", func(t *testing.T) {
		m := &AddonManifest{Header: Header{UUID: "uuid"}}
		_, err := IdentifyPackKind(m)
		require.Error(t, err, "expected error for manifest with no modules, got nil")
	})

	t.Run("unrecognized module type", func(t *testing.T) {
		m := &AddonManifest{
			Header:  Header{UUID: "uuid"},
			Modules: []Module{{Type: "something_else"}},
		}
		_, err := IdentifyPackKind(m)
		require.Error(t, err, "expected error for unrecognized module type, got nil")
	})

	t.Run("nil manifest", func(t *testing.T) {
		_, err := IdentifyPackKind(nil)
		require.Error(t, err, "expected error for nil manifest, got nil")
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
		require.NoError(t, err)
		require.Equal(t, BehaviorPack, kind)
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
		require.Equal(t, want, got)
	}
}
