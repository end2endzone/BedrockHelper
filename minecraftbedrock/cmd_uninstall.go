package minecraftbedrock

import "fmt"

// UninstallAddonInServer uninstalls every pack contained in the add-on file at the given addonPath
// from the Minecraft Bedrock server installed at serverDir.
// It returns the list of packs that were uninstalled or an error.
func UninstallAddonInServer(addonPath, serverDir string) ([]*Pack, error) {
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

	// Proceed with the uninstall of the addon
	uninstalledPacks, err := activeWorld.UninstallAddon(addonPath)
	if err != nil {
		return nil, err
	}

	return uninstalledPacks, nil
}

// UninstallPackInServerByUUID uninstalls a single pack, identified by a UUID
// from the Minecraft Bedrock server installed at serverDir.
// This function is useful when the original add-on file is no longer available or has been deleted.
func UninstallPackInServerByUUID(uuid, serverDir string) (*Pack, error) {
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

	// Find the specific pack to uninstall
	pack := FindPackByUUID(packs, uuid)
	if pack == nil {
		return nil, fmt.Errorf("no pack installed with UUID: %q", uuid)
	}

	// Do the uninstall
	uninstalledPack, err := activeWorld.UninstallPack(pack)

	return uninstalledPack, nil
}
