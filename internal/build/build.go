package build

import (
	"fmt"
	"runtime/debug"
)

func GetBuildTagFromMetadata() (string, error) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		err := fmt.Errorf("build-info metadata not available")
		return "", err
	}

	// When compiling using `go install URL@tag` the `info.Main.Version` variable contains the name of the tag used.
	// For example, when running `go install github.com/username/reponame/cmd/my-app@v0.2.0-alpha` Go's toolchain/compiler
	// will set info.Main.Version to `v0.2.0-alpha`.

	// When using `latest` for tag, Go's install command will find the latest tag that is not `alpha`, `beta`, etc.
	// For example, when the following tag exists:
	// - v0.2.0
	// - v0.3.0
	// - v0.3.1-alpha
	// - v0.3.1-beta
	// Go's toolchain/compiler will resolve the latest version as v0.3.0 and will set info.Main.Version to `v0.3.0`.

	return info.Main.Version, nil
}
