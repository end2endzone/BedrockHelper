package libbedrockpacks

import (
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindAddonsInDir_NonRecursive(t *testing.T) {
	dir := testdataDir(t) + "/addons"
	got, err := FindAddonsInDir(dir, false)
	require.NoError(t, err)

	names := baseNames(got)
	sort.Strings(names)

	// File corrupted.mcaddon has the right extension but is not a valid zip, so it
	// must be excluded. zip_with_no_manifest.zip is a valid zip (even without a
	// manifest) so it counts as a discoverable add-on file.
	want := []string{
		"behavior_only.mcpack",
		"foobar.mcaddon",
		"zip_with_no_manifest.zip",
		"solo.mcpack"}
	sort.Strings(want)

	require.Equal(t, len(want), len(names))
	for i := range names {
		require.Equal(t, want[i], names[i], "failure at index %v", i)
	}
}

func TestFindAddonsInDir_Recursive(t *testing.T) {
	dir := testdataDir(t)

	got, err := FindAddonsInDir(dir, true)
	require.NoError(t, err)

	names := baseNames(got)
	sort.Strings(names)

	want := []string{
		"foobar.mcaddon",
		"solo.mcpack",
		"behavior_only.mcpack",
		"zip_with_no_manifest.zip"}
	sort.Strings(want)

	require.Equal(t, len(want), len(names), "want={%v},  got={%v}", want, names)
	for i := range names {
		require.Equal(t, want[i], names[i], "failure at index %v", i)
	}
}

func TestFindAddonsInDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := FindAddonsInDir(dir, true)
	require.NoError(t, err)
	require.Equal(t, 0, len(got))
}

func baseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}
