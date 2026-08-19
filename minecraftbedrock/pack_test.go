package minecraftbedrock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// makeTestPack is a utility function to easily create a temporary pack from data for tests
func makeTestPack(path string, name string, kind PackKind, uuid string, version Version) *Pack {
	pack := &Pack{
		Path: path,
		Manifest: &AddonManifest{
			Header: Header{
				Name:    name,
				UUID:    uuid,
				Version: version,
			},
		},
	}

	switch kind {
	case BehaviorPack:
		pack.Manifest.Modules = []Module{
			{
				Type: "data",
			},
			{
				Type: "string",
			},
		}
	case ResourcePack:
		pack.Manifest.Modules = []Module{
			{
				Type: "resources",
			},
		}
	default:
	}

	return pack
}

// loadPackFixture extracts the given testdata addon fixture into a temporary directory and
// loads the pack found at its root (i.e. a standalone .mcpack-style layout).
// The function does not support addons because LoadPackFromDirectory() will fail to find a manifest.json in root directory.
func loadPackFixture(t *testing.T, name string) *Pack {
	t.Helper()
	extractDir := t.TempDir()
	require.NoError(t, ExtractZip(getAddonFixturePath(t, name), extractDir))
	pack, err := LoadPackFromDirectory(extractDir)
	require.NoError(t, err)
	return pack
}

func TestPack_Kind(t *testing.T) {
	t.Run("behavior pack", func(t *testing.T) {
		pack := loadPackFixture(t, "behavior_only.mcpack")

		kind, err := pack.Kind()
		require.NoError(t, err)

		// Assert
		require.Equal(t, BehaviorPack, kind)
		require.Equal(t, BehaviorPack, pack.KindSafe())
	})

	t.Run("resource pack", func(t *testing.T) {
		pack := loadPackFixture(t, "solo.mcpack")

		kind, err := pack.Kind()
		require.NoError(t, err)

		// Assert
		require.Equal(t, ResourcePack, kind)
		require.Equal(t, ResourcePack, pack.KindSafe())
	})

	t.Run("unidentifiable manifest returns an error from Kind and UnknownPack from KindSafe", func(t *testing.T) {
		pack := Pack{
			Path: "/tmp/does-not-matter",
			Manifest: &AddonManifest{
				Header: Header{
					Name: "Weird Pack",
					UUID: "cccccccc-cccc-cccc-cccc-cccccccccccc"},
				// No modules in header --> IdentifyPackKind() cannot classify this manifest.
			},
		}

		_, err := pack.Kind()
		require.Error(t, err)

		// Assert
		require.Equal(t, UnknownPack, pack.KindSafe())
	})
}

func TestPack_NameAndUUID(t *testing.T) {
	pack := loadPackFixture(t, "solo.mcpack")
	require.Equal(t, "Solo RP", pack.Name())
	require.Equal(t, "66666666-6666-6666-6666-666666666666", pack.UUID())
}

func TestPack_NameSanitized(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		{"Solo RP", "Solo RP"},
		{"Weird/Name:Here", "Weird_Name_Here"},
	}

	for _, tc := range cases {
		// Create a fake pack whose name matches test values
		pack := Pack{
			Path: "/tmp/does-not-matter",
			Manifest: &AddonManifest{
				Header: Header{
					Name: tc.name,
				},
			},
		}

		// Assert
		require.Equal(t, tc.want, pack.NameSanitized())
	}

	t.Run("empty name", func(t *testing.T) {
		pack := Pack{
			Path: "/tmp/does-not-matter",
			Manifest: &AddonManifest{
				Header: Header{
					Name: "",
				},
			},
		}

		// Act
		sanitizedName := pack.NameSanitized()

		// Assert
		require.True(t, StringsCompareN("pack", sanitizedName, 4))
		require.Greater(t, len(sanitizedName), 4)
	})

}

func TestPack_Description(t *testing.T) {
	pack := loadPackFixture(t, "solo.mcpack")
	desc := pack.Description()

	// Assert description shall contains the following values...
	require.Contains(t, desc, pack.Name())
	require.Contains(t, desc, pack.UUID())
	require.Contains(t, desc, pack.KindSafe().String())
}

func TestLoadPackFromDirectory(t *testing.T) {
	t.Run("valid pack directory", func(t *testing.T) {
		extractDir := t.TempDir()
		require.NoError(t, ExtractZip(getAddonFixturePath(t, "solo.mcpack"), extractDir))
		pack, err := LoadPackFromDirectory(extractDir)
		require.NoError(t, err)

		// Assert
		require.Equal(t, "Solo RP", pack.Name())
	})

	t.Run("directory without a manifest.json", func(t *testing.T) {
		dir := t.TempDir()
		_, err := LoadPackFromDirectory(dir)
		require.Error(t, err)
	})

	t.Run("not compatible with addons", func(t *testing.T) {
		extractDir := t.TempDir()
		require.NoError(t, ExtractZip(getAddonFixturePath(t, "foobar.mcaddon"), extractDir))
		_, err := LoadPackFromDirectory(extractDir)

		// Assert
		require.Error(t, err)
	})
}

func TestLoadPacksFromSubdirectories(t *testing.T) {
	t.Run("valid addon directory", func(t *testing.T) {
		extractDir := t.TempDir()
		require.NoError(t, ExtractZip(getAddonFixturePath(t, "foobar.mcaddon"), extractDir))

		packs, err := LoadPacksFromSubdirectories(extractDir)
		require.NoError(t, err)
		require.Len(t, packs, 2)

		// Assert
		actualNames := []string{packs[0].Name(), packs[1].Name()}
		expectedNames := []string{"Foobar BP", "Foobar RP"}
		require.ElementsMatch(t, expectedNames, actualNames)
	})

	t.Run("not compatible with pack", func(t *testing.T) {
		extractDir := t.TempDir()
		require.NoError(t, ExtractZip(getAddonFixturePath(t, "solo.mcpack"), extractDir))

		_, err := LoadPacksFromSubdirectories(extractDir)

		// Assert
		require.Error(t, err)
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		_, err := LoadPacksFromSubdirectories("/tmp/does-not-exist-xyz")
		require.Error(t, err)
	})
}

func TestLoadAllPacksFromDirectoriesOrSubdirectories(t *testing.T) {
	t.Run("valid addon directory", func(t *testing.T) {
		extractDir := t.TempDir()
		require.NoError(t, ExtractZip(getAddonFixturePath(t, "foobar.mcaddon"), extractDir))

		packs, err := LoadAllPacksFromDirectoriesOrSubdirectories(extractDir)
		require.NoError(t, err)
		require.Len(t, packs, 2)

		// Assert
		actualNames := []string{packs[0].Name(), packs[1].Name()}
		expectedNames := []string{"Foobar BP", "Foobar RP"}
		require.ElementsMatch(t, expectedNames, actualNames)
	})

	t.Run("valid pack directory", func(t *testing.T) {
		extractDir := t.TempDir()
		require.NoError(t, ExtractZip(getAddonFixturePath(t, "solo.mcpack"), extractDir))

		packs, err := LoadAllPacksFromDirectoriesOrSubdirectories(extractDir)
		require.NoError(t, err)
		require.Len(t, packs, 1)

		// Assert
		require.Equal(t, "Solo RP", packs[0].Name())
	})

	t.Run("empty directory report no error", func(t *testing.T) {
		packs, err := LoadAllPacksFromDirectoriesOrSubdirectories(t.TempDir())
		require.NoError(t, err)
		require.Empty(t, packs)
	})
}

func TestFindPackByUUID(t *testing.T) {
	pack1 := loadPackFixture(t, "behavior_only.mcpack")
	pack2 := loadPackFixture(t, "solo.mcpack")
	packs := []*Pack{pack1, pack2}

	found := FindPackByUUID(packs, pack2.UUID())
	require.NotNil(t, found)
	require.Equal(t, pack2.Name(), found.Name())

	require.Nil(t, FindPackByUUID(packs, "00000000-0000-0000-0000-000000000000"))
}

func TestFilterPacksByUUID(t *testing.T) {
	pack1 := loadPackFixture(t, "behavior_only.mcpack")
	pack2 := loadPackFixture(t, "solo.mcpack")
	packs := []*Pack{pack1, pack2}

	filtered := FilterPacksByUUID(packs, pack1.UUID())
	require.Len(t, filtered, 1)
	require.Equal(t, pack1.Name(), filtered[0].Name()) // compare by name

	require.Empty(t, FilterPacksByUUID(packs, "00000000-0000-0000-0000-000000000000"))
}

func TestFilterPacksByKind(t *testing.T) {
	pack1 := loadPackFixture(t, "behavior_only.mcpack")
	pack2 := loadPackFixture(t, "solo.mcpack")
	packs := []*Pack{pack1, pack2}

	behaviorOnly := FilterPacksByKind(packs, BehaviorPack)
	resourceOnly := FilterPacksByKind(packs, ResourcePack)

	// Assert
	require.Len(t, behaviorOnly, 1)
	require.Equal(t, pack1.Name(), behaviorOnly[0].Name())

	require.Len(t, resourceOnly, 1)
	require.Equal(t, pack2.Name(), resourceOnly[0].Name())
}

func TestRemoveFormattingInPackName(t *testing.T) {
	cases := []struct {
		fancyName string
		expected  string
	}{
		{"§6orange text§r", "orange text"},
		{"§bblue text§r", "blue text"},
		{"normal name", "normal name"},
		{"random normal text §7grey text§5purple text§r", "random normal text grey textpurple text"},
		{"normal §cred text§fwhite text", "normal red textwhite text"},
		{"§l§7grey text §rfooter", "grey text footer"},
	}

	for _, tc := range cases {
		actual := RemoveFormattingInPackName(tc.fancyName)

		// Assert
		require.Equal(t, tc.expected, actual)
	}
}
