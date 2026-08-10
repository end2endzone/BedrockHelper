package libminecraftbedrock

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

	// Get the list of addons
	// File corrupted.mcaddon has the right extension but is not a valid zip
	want := []string{
		"behavior_only.mcpack",
		"foobar.mcaddon",
		"solo.mcpack",
		"zip_with_no_manifest.zip", // is a valid zip (without a manifest) so it counts as a discoverable add-on file
	}
	sort.Strings(want)

	// Assert both list are equals
	require.ElementsMatch(t, want, names)
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

	// Assert both list are equals
	require.ElementsMatch(t, want, names)
}

func TestFindAddonsInDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := FindAddonsInDir(dir, true)
	require.NoError(t, err)
	require.Empty(t, got)
}

func baseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}
