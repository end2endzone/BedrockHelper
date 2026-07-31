package libbedrockpacks

// ListInstalledPacks lists the packs currently registered for a given Minecraft Bedrock server located at serverDir.
// For each registered UUID, it attempts to resolve the name of the pack by scanning the corresponding behavior_packs/ or
// resource_packs/ directories for a matching manifest.json.
func ListInstalledPacks(serverDir string) ([]*Pack, error) {
	// Get the server
	server, err := GetServer(serverDir)
	if err != nil {
		return nil, err
	}

	// Get the active world
	activeWorld, err := server.ActiveWorld()
	if err != nil {
		return nil, err
	}

	// Find all packs in world
	packs, err := activeWorld.Packs()
	if err != nil {
		return nil, err
	}

	return packs, nil
}
