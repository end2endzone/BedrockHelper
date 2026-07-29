package libbedrockpacks

import "fmt"

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

// Version represents a Minecraft Bedrock manifest [major, minor, patch] version triplet.
type Version [3]int

// String renders the version as "major.minor.patch".
func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v[0], v[1], v[2])
}

// InstalledPack describes a pack that was installed, discovered or resolved to its on-disk location.
type InstalledPack struct {
	UUID      string
	Name      string
	Kind      PackKind
	Version   Version
	Directory string // for packs inside an addon or installed in a server
}

// RegisteredPack describes an entry found in a world registry file.
type RegisteredPack struct {
	UUID    string
	Name    string // optionnal
	Kind    PackKind
	Version Version
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
