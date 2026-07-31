package libbedrockpacks

// ListInstalledPacks lists the packs currently registered for a given Minecraft Bedrock server located at serverDir.
// For each registered UUID, it attempts to resolve the name of the pack by scanning the corresponding behavior_packs/ or
// resource_packs/ directories for a matching manifest.json.
func ListInstalledPacks(serverDir string) ([]RegisteredPack, error) {
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

	// For each pack, resolve its properties (name, uuid, etc)
	var results []RegisteredPack
	for _, pack := range packs {

		registeredPack := RegisteredPack{
			UUID:    pack.Manifest.Header.UUID,
			Name:    pack.Name(),
			Kind:    pack.KindSafe(),
			Version: pack.Manifest.Header.Version,
		}

		// add to our results
		results = append(results, registeredPack)
	}

	return results, nil
}
