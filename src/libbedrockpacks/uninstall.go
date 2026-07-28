package libbedrockpacks

// UninstallAddon uninstalls every pack contained in the add-on file at the given addonPath
// from the Minecraft Bedrock server installed at serverDir.
// It returns the list of packs that were uninstalled or an error.
func UninstallAddon(addonPath, serverDir string) ([]InstalledPack, error) {
	return nil, NotImplementedErr()
}

// UninstallPackByUUID uninstalls a single pack, identified by a UUID
// from the Minecraft Bedrock server installed at serverDir.
// This function is useful when the original add-on file is no longer available or has been deleted.
func UninstallPackByUUID(uuid, serverDir string) (InstalledPack, error) {
	if ok, err := IsServerDirectory(serverDir); !ok {
		return InstalledPack{}, err
	}

	_ /*activeWorldDir*/, err := FindActiveWorldDir(serverDir)
	if err != nil {
		return InstalledPack{}, err
	}

	return InstalledPack{}, NotImplementedErr()
}
