package libbedrockpacks

// FindAddonsInDir searches dir (and its subdirectories, if recursive is
// true) for files that are valid add-on packs (.zip, .mcaddon, .mcpack that
// are readable as zip archives). It returns their absolute paths.
func FindAddonsInDir(dir string, recursive bool) ([]string, error) {
	return nil, NotImplementedErr()
}
