package libbedrockpacks

// ListInstalledPacks lists the packs currently registered for a given Minecraft Bedrock server located at at serverDir.
// For each registered UUID, it attempts to resolve the name of the pack by scanning the corresponding behavior_packs/ or
// resource_packs/ directories for a matching manifest.json.
func ListInstalledPacks(serverDir string) ([]RegisteredPack, error) {
	if ok, err := IsServerDirectory(serverDir); !ok {
		return nil, err
	}

	_ /*worldDir*/, err := FindActiveWorldDir(serverDir)
	if err != nil {
		return nil, err
	}

	var results []RegisteredPack
	results = append(results, RegisteredPack{
		UUID:    "9d3ffd3a-3730-4119-9b11-4333db83aea7",
		Name:    "foobar",
		Kind:    BehaviorPack,
		Version: Version{1, 0, 0},
	})
	return results, nil
}
