package main

import (
	"fmt"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

// These variables are populated at build time using -ldflags
var (
	Version    = "0.0.0"
	CommitHash = "0000000"
	BuildDate  = "1900-01-01"
)

type ProductVersion struct {
	Version    string
	CommitHash string
	BuildDate  string
}

// isNumberInRange validate that the given input str is a number in the range [min;max]
// Returns true when the input is in range. Returns false otherwise.
func isNumberInRange(str string, min int, max int) bool {
	num, err := strconv.Atoi(str)
	if err != nil {
		// Parsing error
		return false
	}

	if num < min || num > max {
		return false
	}
	return true
}

// isTimeStamp detect if the given string is a timestamp. A timestamp is a string in format `YYYYMMDDhhmmss`.
// Returns true if the given input is a timestamp. Returns false otherwise.
func isTimeStamp(input string) bool {
	if len(input) != 14 {
		return false
	}

	if input[0:2] != "20" {
		// Not from 2000 era.
		return false
	}

	year := input[0:4]
	month := input[4:6]
	day := input[6:8]
	hours := input[8:10]
	minutes := input[10:12]
	seconds := input[12:14]

	if !isNumberInRange(year, 2000, 2100) ||
		!isNumberInRange(month, 1, 12) ||
		!isNumberInRange(day, 1, 31) ||
		!isNumberInRange(hours, 1, 24) ||
		!isNumberInRange(minutes, 1, 59) ||
		!isNumberInRange(seconds, 1, 59) {
		return false
	}
	return true
}

// isGitHash detect if the given string is a git hash.
// Returns true if the given input is a git hash. Returns false otherwise.
func isGitHash(input string) bool {
	for _, c := range input {
		valid := (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
		if !valid {
			return false
		}
	}
	return true
}

// parseDateTimeAndGitHash parse a datetime and a git hash from the input string.
// Returns  ok when the input string is a semantic prerelease string containing a datetime and a git hash.
// Returns !ok otherwise
func parseDateTimeAndGitHash(input string) (datetime string, gitcommit string, ok bool) {
	// example: -20260815131415-de3798c09c08

	// Remove leading `-` character
	input = strings.TrimPrefix(input, "-")

	fields := strings.Split(input, "-")
	if len(fields) != 2 {
		return "", "", false
	}

	// Check first field
	if !isTimeStamp(fields[0]) {
		return "", "", false
	}

	// Check second field
	if !isGitHash(fields[1]) {
		return "", "", false
	}

	return fields[0], fields[1], true
}

// GetSemanticVersionFromDebugBuildInfo parses a go pseudo-version into a ProductVersion.
// A pseudo-version is a string in format "v1.2.3-20260815131415-de3798c09c08+dirty".
func parsePseudoVersion(input string) ProductVersion {
	p := ProductVersion{}

	// semver.Canonical removes the "+dirty" and standardizes it
	canonical := semver.Canonical(input)

	// Strip the leading "v" if you strictly want the digits and pre-release labels
	canonical = strings.TrimPrefix(canonical, "v")

	p.Version = canonical

	// Can we extract datetime and githash from pre-release labels ?
	prerelease := semver.Prerelease(input)
	datetime, githash, ok := parseDateTimeAndGitHash(prerelease)
	if ok {
		p.Version = strings.ReplaceAll(canonical, prerelease, "") // remove pre-release string from canonical version
		p.BuildDate = datetime[0:4] + "-" + datetime[4:6] + "-" + datetime[6:8]
		p.CommitHash = githash
	}

	return p
}

func GetProductVersion() ProductVersion {
	// Create a product version from local variables.
	p := ProductVersion{
		Version:    Version,
		CommitHash: CommitHash,
		BuildDate:  BuildDate,
	}

	// If the product version was set with the legacy `-ldflags` command line at build time
	if p.Version != "0.0.0" {
		return p
	}

	// Try to parse values which are automatically embedded by Go toolchain
	info, ok := debug.ReadBuildInfo()
	if !ok {
		// Can't do better
		return p
	}

	// Git Hash and Commit Date: Automatically embedded by Go toolchain
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			p.CommitHash = setting.Value[0:11]
		case "vcs.time":
			p.BuildDate = setting.Value

			// Example value: `2026-08-15T13:14:15Z`. Try to truncate for keeping only the date

			// Parse the string directly into a time.Time object
			buildTime, err := time.Parse(time.RFC3339, setting.Value)
			if err != nil {
				continue // keep p.BuildDate as is
			}

			year, month, day := buildTime.Date()
			p.BuildDate = fmt.Sprintf("%d-%02d-%02d", year, month, day)
		}
	}

	// Version: if installed via `go install` tagged release (for example `@v1.2.3`)
	// Warning: version can default to `(devel)` if run locally or built without version flags.
	// For example, if you run your app using `go run` . or compile it locally without tags, Go sets info.Main.Version to the string "(devel)".
	if info.Main.Version != "" {
		p.Version = info.Main.Version
	}

	if p.Version == "(devel)" {
		// Can't do better
		return p
	}

	// Assume info.Main.Version is in format `v1.2.3-20260815131415-de3798c09c08+dirty`.
	pseudo := parsePseudoVersion(info.Main.Version)
	p.Version = pseudo.Version

	return p
}

func GetProductVersionString() string {
	p := GetProductVersion()

	msg := fmt.Sprintf("%s (%s) compiled on %s", p.Version, p.CommitHash, p.BuildDate)
	return msg
}
