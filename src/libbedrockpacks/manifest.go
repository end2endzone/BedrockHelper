package libbedrockpacks

import (
	"encoding/json"
	"fmt"
	"os"
)

// LoadManifestFromBytes parses the raw JSON bytes of a manifest.json file as an AddonManifest structure.
func LoadManifestFromBytes(data []byte) (*AddonManifest, error) {
	var m AddonManifest
	err := json.Unmarshal(data, &m)
	if err != nil {
		return nil, fmt.Errorf("failed to parse manifest.json: %w", err)
	}
	if m.Header.UUID == "" {
		return nil, fmt.Errorf("manifest.json is missing a header.uuid field")
	}
	return &m, nil
}

// LoadManifestFromFile loads a manifest.json data from a file path as an AddonManifest structure.
func LoadManifestFromFile(path string) (*AddonManifest, error) {
	// Read the manifest's json file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q as a manifest: %w", path, err)
	}

	// Parse it as a AddonManifest pointer
	var manifest *AddonManifest
	manifest, err = LoadManifestFromBytes(data)

	return manifest, err
}

// IdentifyPackKind inspects a AddonManifest sutrct to determines if the manifest matches a behavior pack or a resource pack.
// Behavior packs use module types "data" or "script".
// Resource packs use module types "resources", "client_data" or "interface".
func IdentifyPackKind(m *AddonManifest) (PackKind, error) {
	if m == nil {
		return UnknownPack, fmt.Errorf("nil manifest")
	}
	if len(m.Modules) == 0 {
		return UnknownPack, fmt.Errorf("manifest %s has no modules to identify its pack kind", m.Header.UUID)
	}

	haveBehavior := false
	haveResource := false

	for _, mod := range m.Modules {
		switch mod.Type {
		case "data", "script":
			haveBehavior = true
		case "resources", "client_data", "interface":
			haveResource = true
		}
	}

	switch {
	case haveBehavior && !haveResource:
		return BehaviorPack, nil
	case haveResource && !haveBehavior:
		return ResourcePack, nil
	case haveBehavior && haveResource:
		// Mixed-module packs are rare; behavior takes precedence since a
		// "data"/"script" module means the world must load its scripting.
		// Also behavior packs have dependencies to resources packs and not the other way around.
		return BehaviorPack, nil
	default:
		return UnknownPack, fmt.Errorf("manifest %s has an unrecognized module type", m.Header.UUID)
	}
}
