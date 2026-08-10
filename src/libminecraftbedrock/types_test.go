package libminecraftbedrock

import (
	"testing"

	"github.com/stretchr/testify/require"
)

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

func TestPackKindRegistryFileName(t *testing.T) {
	behaviorName, err := BehaviorPack.RegistryFileName()
	require.NoError(t, err)
	require.Equal(t, "world_behavior_packs.json", behaviorName)

	resourceName, err := ResourcePack.RegistryFileName()
	require.NoError(t, err)
	require.Equal(t, "world_resource_packs.json", resourceName)

	_, err = UnknownPack.RegistryFileName()
	require.Error(t, err)
}

func TestPackKindInstallDirName(t *testing.T) {
	behaviorDir, err := BehaviorPack.InstallDirName()
	require.NoError(t, err)
	require.Equal(t, "behavior_packs", behaviorDir)

	resourceDir, err := ResourcePack.InstallDirName()
	require.NoError(t, err)
	require.Equal(t, "resource_packs", resourceDir)

	_, err = UnknownPack.InstallDirName()
	require.Error(t, err)
}
