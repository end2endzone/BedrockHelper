package version

import (
	"testing"

	_ "embed"

	"github.com/stretchr/testify/require"
)

//go:embed VERSION
var versionFileContent string

func TestGetProductVersionStringContainsVERSIONFile(t *testing.T) {
	productVersion := GetProductVersionString()

	// Assert the string returned from GetProductVersionString() contains the actual VERSION string.
	require.Contains(t, productVersion, versionFileContent, "GetProductVersionString() does not contains the VERSION. The VERSION file might be outdated.")
}
