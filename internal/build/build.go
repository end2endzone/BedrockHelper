package build

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
)

func PrintBuildInfoMetadata() {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		fmt.Fprintf(os.Stderr, "build-info-metadata not available\n")
		return
	}

	// Add detailed version string
	fmt.Fprintf(os.Stdout, "info.Main.Version=%s\n", info.Main.Version)
	fmt.Fprintf(os.Stdout, "info.Path=%s\n", info.Path)
	fmt.Fprintf(os.Stdout, "info.Main.Path=%s\n", info.Main.Path)

	// Serialize info.BuildSettings to json and add to msg
	jsonData, err := json.MarshalIndent(info.Settings, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	fmt.Fprintf(os.Stdout, "info.Settings=%s\n", string(jsonData))

	// Serialize info.Deps to json and add to msg
	jsonData, err = json.MarshalIndent(info.Deps, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return
	}
	fmt.Fprintf(os.Stdout, "info.Deps=%s\n", string(jsonData))
}
