package minecraftbedrock

// InstallAddonInServer installs every pack contained in the given add-on file (addonPath)
// into the given Minecraft Bedrock server located at serverDir.
// Returns the list of packs that were installed.
// Returns an error otherwise.
func InstallAddonInServer(addonPath string, serverDir string) ([]*Pack, error) {
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

	// Proceed with the install of the addon
	installedPacks, err := activeWorld.InstallAddon(addonPath)
	if err != nil {
		return nil, err
	}

	return installedPacks, nil
}
