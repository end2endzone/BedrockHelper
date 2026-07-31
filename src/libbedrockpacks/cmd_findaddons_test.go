package libbedrockpacks

import (
	"path/filepath"
	"sort"
	"testing"
)

func TestFindAddonsInDir_NonRecursive(t *testing.T) {
	dir := testdataDir(t) + "/addons"
	got, err := FindAddonsInDir(dir, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

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
	if !equalStrings(names, want) {
		t.Errorf("FindAddonsInDir(non-recursive) = %v, want %v", names, want)
	}
}

func TestFindAddonsInDir_Recursive(t *testing.T) {
	dir := testdataDir(t)
	got, err := FindAddonsInDir(dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	names := baseNames(got)
	found := make(map[string]bool)
	for _, n := range names {
		found[n] = true
	}
	for _, want := range []string{"foobar.mcaddon", "solo.mcpack", "behavior_only.mcpack"} {
		if !found[want] {
			t.Errorf("expected recursive scan to find %q, results were: %v", want, names)
		}
	}
}

func TestFindAddonsInDir_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	got, err := FindAddonsInDir(dir, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Errorf("expected no add-ons in an empty directory, got %v", got)
	}
}

func baseNames(paths []string) []string {
	out := make([]string, len(paths))
	for i, p := range paths {
		out[i] = filepath.Base(p)
	}
	return out
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	sort.Strings(a)
	sort.Strings(b)
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
