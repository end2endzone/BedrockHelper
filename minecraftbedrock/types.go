package minecraftbedrock

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// PackKind identifies whether a pack found inside an add-on is a behavior or a resource pack.
type PackKind int

const (
	// UnknownPack is returned when the pack kind could not be determined from its manifest.
	UnknownPack PackKind = iota

	// BehaviorPack identifies a Minecraft Bedrock behavior pack.
	BehaviorPack

	// ResourcePack identifies a Minecraft Bedrock resource pack.
	ResourcePack
)

// AllPackKinds lists all the different kinds of Pack.
var AllPackKinds = []PackKind{BehaviorPack, ResourcePack}

// String implements fmt.Stringer for PackKind.
func (k PackKind) String() string {
	switch k {
	case BehaviorPack:
		return "BehaviorPack"
	case ResourcePack:
		return "ResourcePack"
	default:
		return "UnknownPack"
	}
}

// RegistryFileName returns the world registry file name that manages the given kind of pack.
// Possible return values are `world_behavior_packs.json`, `world_resource_packs.json` or an error.
func (k PackKind) RegistryFileName() (string, error) {
	switch k {
	case BehaviorPack:
		return "world_behavior_packs.json", nil
	case ResourcePack:
		return "world_resource_packs.json", nil
	default:
		return "", fmt.Errorf("cannot determine registry file for unknown pack kind")
	}
}

// InstallDirName returns the directory name where a pack of the given kind must be installed in a world.
// Possible return values are `behavior_packs`, `resource_packs` or an error.
func (k PackKind) InstallDirName() (string, error) {
	switch k {
	case BehaviorPack:
		return "behavior_packs", nil
	case ResourcePack:
		return "resource_packs", nil
	default:
		return "", fmt.Errorf("cannot determine install directory for unknown pack kind")
	}
}

type Command int

const (
	Install   Command = iota // 0
	Uninstall                // 1
)

func (command Command) String() string {
	switch command {
	case Install:
		return "Install"
	case Uninstall:
		return "Uninstall"
	default:
		return "Unknown"
	}
}

// Version represents a Minecraft Bedrock manifest [major, minor, patch] version triplet.
type Version [3]int

// String renders the version as "major.minor.patch".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}

// UnmarshalJSON unmarshal a Version from an array of 3 integers (`[1, 2, 3]`) or from a string (`"4.5.6"`).
// Some manifest specifies versions as string instead of the offical 3 integer array.
func (v *Version) UnmarshalJSON(data []byte) error {
	// First, try unmarshaling it as the standard [3]int array
	var arrayFormat [3]int
	if err := json.Unmarshal(data, &arrayFormat); err == nil {
		*v = Version(arrayFormat)
		return nil
	}

	// If that fails, handle the string format (for example, "1.2.3")
	var stringFormat string
	err := json.Unmarshal(data, &stringFormat)
	if err != nil {
		return err
	}

	// Parse the major, minor, and patch values out of the string
	var major, minor, patch int
	parts := strings.Split(stringFormat, ".")
	for i := 0; i < 3; i++ {
		// If the string doesn't have this component, fall back to 0
		if i >= len(parts) {
			v[i] = 0
			continue
		}

		// Convert the string segment to an integer
		val, err := strconv.Atoi(parts[i])
		if err != nil {
			return fmt.Errorf("failed to parse version '%s': %w", stringFormat, err)
		}

		v[i] = val
	}

	// Assign the parsed integers back to the original array pointer
	*v = Version([3]int{major, minor, patch})
	return nil
}

// Header mirrors the "header" object of a Minecraft Bedrock manifest.json.
type Header struct {
	Name             string  `json:"name"`
	Description      string  `json:"description,omitempty"`
	UUID             string  `json:"uuid"`
	Version          Version `json:"version"`
	MinEngineVersion Version `json:"min_engine_version,omitempty"`
}

// Module mirrors an entry of the "modules" array of a manifest.json.
type Module struct {
	Type        string  `json:"type"`
	UUID        string  `json:"uuid,omitempty"`
	Version     Version `json:"version"`
	Description string  `json:"description,omitempty"`
	Entry       string  `json:"entry,omitempty"`
}

// Dependency mirrors an entry of the "dependencies" array of a manifest.json.
type Dependency struct {
	UUID       string  `json:"uuid,omitempty"`
	ModuleName string  `json:"module_name,omitempty"`
	Version    Version `json:"version,omitempty"`
}

// Metadata mirrors the optional "metadata" object of a manifest.json.
type Metadata struct {
	Authors []string `json:"authors,omitempty"`
	License string   `json:"license,omitempty"`
	URL     string   `json:"url,omitempty"`
}

// AddonManifest mirrors the structure of a Minecraft Bedrock pack manifest.json file.
type AddonManifest struct {
	FormatVersion int          `json:"format_version"`
	Header        Header       `json:"header"`
	Modules       []Module     `json:"modules,omitempty"`
	Dependencies  []Dependency `json:"dependencies,omitempty"`
	Capabilities  []string     `json:"capabilities,omitempty"`
	Metadata      *Metadata    `json:"metadata,omitempty"`
}
