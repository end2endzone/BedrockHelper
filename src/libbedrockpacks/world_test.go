package libbedrockpacks

import (
	"path/filepath"
	"testing"
)

func TestFindActiveWorldDir(t *testing.T) {
	cases := []struct {
		testName     string
		fixture      string
		expectSuffix string
		expectError  bool
	}{
		{"Test random directory", "not_a_server", "", true},
		{"Test missing bedrock_server.exe", "not_a_server_missing_exec", "", true},
		{"Test no server.properties", "not_a_server_missing_server.properties", "", true},
		{"Test no worlds directory", "not_a_server_missing_worlds", "", true},

		{"Test server with content", "server", "worlds/Bedrock level", false},
		{"Test missing level-name falls back to first world dir", "server_no_level_name", "worlds/MyWorld", false},
		{"Test full server", "server_with_installed_pack", "worlds/Bedrock level", false},
	}

	for _, tc := range cases {
		t.Run(tc.testName, func(t *testing.T) {
			serverDir := getServerFixturePath(t, tc.fixture)
			got, err := FindActiveWorldDir(serverDir)

			if !tc.expectError && err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tc.expectError && err == nil {
				t.Fatalf("expected an error, got nil")
			}

			if !tc.expectError {
				// FindActiveWorldDir() returns absolute paths
				// Build an absolute path from the expected relative path

				want := serverDir + "/" + tc.expectSuffix

				want = filepath.Clean(want) // normalize expected path to match the file separator on the system

				if got != want {
					t.Errorf("FindActiveWorldDir(%q) = %q, want %q", serverDir, got, want)
				}
			}

		})
	}
}
