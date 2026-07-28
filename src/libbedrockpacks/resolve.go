package libbedrockpacks

import (
	"fmt"
)

// ResolvePackByUUID searches serverDir (recursively) for an add-on file containing a pack whose manifest UUID matches the given UUID.
// It returns the path of the first matching add-on file found.
func ResolvePackByUUID(uuid string, serverDir string) (string, error) {
	ok, err := IsServerDirectory(serverDir)
	if !ok || err != nil {
		return "", err
	}

	_ /*addonPaths*/, err = FindAddonsInDir(serverDir, true)
	if err != nil {
		return "", err
	}

	//for _, _addonPath := range addonPaths {
	//}

	return "", fmt.Errorf("no add-on file under %q contains a pack with uuid %s", serverDir, uuid)
}
