package main

import (
	"testing"

	root "github.com/end2endzone/BedrockHelper"

	"github.com/stretchr/testify/require"
)

func TestGetProductVersionStringContainsVERSIONFile(t *testing.T) {
	productVersion := GetProductVersionString()
	versionFileContent := root.GetVersionFromVersionFile()

	// Assert the string returned from GetProductVersionString() contains the actual VERSION string.
	require.Contains(t, productVersion, versionFileContent, "GetProductVersionString() does not contains the VERSION. The VERSION file might be outdated.")
}
