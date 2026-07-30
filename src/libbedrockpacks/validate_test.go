package libbedrockpacks

import "testing"

func TestIsValidServerDirectory(t *testing.T) {
	cases := []struct {
		testName    string
		fixture     string
		expectError bool
		expectValid bool
	}{
		{"Test not a server directory", "not_a_server", true, false},
		{"Test missing executable", "not_a_server_missing_exec", true, false},
		{"Test missing server.properties", "not_a_server_missing_server.properties", true, false},
		{"Test missing worlds directory", "not_a_server_missing_worlds", true, false},

		{"Test server with content", "server", false, true},
		{"Test missing level-name falls back to first world dir", "server_no_level_name", false, true},
		{"Test full server", "server_with_installed_pack", false, true},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			path := getServerFixturePath(t, tc.fixture)
			ok := IsValidServerDirectory(path)
			err := ValidateServerDirectory(path)
			if ok != tc.expectValid {
				t.Fatalf("IsValidServerDirectory(%q) = (%v, %v), expected %v", path, ok, err, tc.expectValid)
			}
			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tc.expectError && err == nil {
				t.Fatalf("expected an error, got nil")
			}
		})
	}

	t.Run("Test nonexistent directory", func(t *testing.T) {
		path := "/path/does/not/exist/at/all"
		ok := IsValidServerDirectory(path)
		err := ValidateServerDirectory(path)
		if ok || err == nil {
			t.Fatalf("expected failure for nonexistent directory, got ok=%v err=%v", ok, err)
		}
	})
}

func TestIsValidAddonFile(t *testing.T) {
	cases := []struct {
		name      string
		file      string
		wantValid bool
	}{
		{"Test valid mcaddon bundle", "foobar.mcaddon", true},
		{"Test valid standalone mcpack", "solo.mcpack", true},
		{"Test valid zip with no manifest (still a valid zip)", "zip_with_no_manifest.zip", true},
		{"Test corrupt file with mcaddon extension", "corrupted.mcaddon", false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := getAddonFixturePath(t, tc.file)
			ok := IsValidAddonFile(path)
			err := ValidateAddonFile(path)
			if ok != tc.wantValid {
				t.Fatalf("IsValidAddonFile(%q) = (%v, %v), want valid=%v", path, ok, err, tc.wantValid)
			}
		})
	}

	t.Run("Test wrong extension", func(t *testing.T) {
		path := getAddonFixturePath(t, "foobar.mcaddon") + ".txt"
		ok := IsValidAddonFile(path)
		err := ValidateAddonFile(path)
		if ok || err == nil {
			t.Fatalf("expected failure for unsupported extension, got ok=%v err=%v", ok, err)
		}
	})

	t.Run("Test nonexistent file", func(t *testing.T) {
		path := "/tmp/does-not-exist-xyz.mcpack"
		ok := IsValidAddonFile(path)
		err := ValidateAddonFile(path)
		if ok || err == nil {
			t.Fatalf("expected failure for nonexistent file, got ok=%v err=%v", ok, err)
		}
	})
}
