package minecraftbedrock

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"

	"github.com/magiconair/properties"
)

type Pack struct {
	Path         string
	Manifest     *AddonManifest
	LanguageList []string
	LanguageMap  map[string]*properties.Properties
}

func (p Pack) Kind() (PackKind, error) {
	kind, err := IdentifyPackKind(p.Manifest)
	if err != nil {
		return UnknownPack, err
	}

	return kind, nil
}

func (p Pack) KindSafe() PackKind {
	kind, err := IdentifyPackKind(p.Manifest)
	if err != nil {
		return UnknownPack
	}

	return kind
}

func (p Pack) Name() string {
	name := p.Manifest.Header.Name

	// Check if the pack don't have localisation properties
	if !p.HasLanguages() {
		return name
	}

	// Resolve the localized name
	nameKey := name

	defaultLangKey := p.GetFirstLocalizedLanguage()
	localizedName, exists := p.GetLocalizedTextValue(defaultLangKey, nameKey)
	if !exists {
		return name
	}

	return localizedName
}

func (p Pack) NameWithoutFormatting() string {
	name := p.Name()
	name = RemoveFormattingInPackName(name)
	return name
}

func (p Pack) NameSanitized() string {
	name := p.Name()

	// Remove formatting such as "§6orange text§r"
	name = RemoveFormattingInPackName(name)

	// Make the content safe for filesystems
	dirName := sanitizeCharactersInPath(name)

	// Prevent empty names
	if dirName == "" {
		// Generate a random name to prevent conflict with other packs with no name.
		randomId := fmt.Sprintf("%05d", rand.N(100000))
		dirName = "pack" + randomId
	}

	return dirName
}

func (p Pack) UUID() string {
	return p.Manifest.Header.UUID
}

func (p Pack) Description() string {
	safeKind, err := p.Kind()
	if err != nil {
		safeKind = UnknownPack
	}

	desc := fmt.Sprintf("%s version %s (%s) uuid=%s", p.NameWithoutFormatting(), p.Manifest.Header.Version, safeKind, p.Manifest.Header.UUID)
	return desc
}

func (p Pack) HasLanguages() bool {
	if len(p.LanguageList) == 0 || len(p.LanguageMap) == 0 {
		return false
	}

	langKey := p.GetFirstLocalizedLanguage()
	if langKey == "" {
		return false
	}

	// Check if language has a property file loaded
	/*value*/
	_, exists := p.LanguageMap[langKey]
	if !exists {
		return false
	}

	return true
}

func (p Pack) GetFirstLocalizedLanguage() string {
	if p.LanguageList == nil || len(p.LanguageList) == 0 {
		return ""
	}
	langKey := p.LanguageList[0]
	return langKey
}

func (p Pack) GetLocalizedTextValue(langKey, textKey string) (value string, exists bool) {
	if p.LanguageList == nil || p.LanguageMap == nil {
		return "", false
	}

	exists = false

	// Get the language's properties
	props, langExists := p.LanguageMap[langKey]
	if !langExists {
		return
	}

	// Get the language's localized text
	propsMap := props.Map()
	text, exists := propsMap[textKey]
	return text, exists
}

func (p Pack) GetDefaultLocalizedTextValue(textKey string) (value string, exists bool) {
	first := p.GetFirstLocalizedLanguage()
	return p.GetLocalizedTextValue(first, textKey)
}

// LoadPackFromDirectory loads a pack stored in the given directory.
// The given directory path must contains a manifest.json file to be a valid pack.
// Returns a valid pack or an error otherwise.
func LoadPackFromDirectory(path string) (*Pack, error) {
	manifestFullPath := filepath.Join(path, "manifest.json")

	// Load its manifest
	manifest, err := LoadManifestFromFile(manifestFullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read a pack from directory %q: %v", path, err)
	}

	// Check for texts/languages.json
	textsDir := filepath.Join(path, "texts")
	languagesFilePath := filepath.Join(textsDir, "languages.json")
	languages, _ := LoadLanguages(languagesFilePath)

	// Create the pack
	pack := &Pack{
		Path:         path,
		Manifest:     manifest,
		LanguageList: languages,
	}

	// Load each language file properties
	for _, langKey := range languages {
		languagePropertiesFileName := fmt.Sprintf("%s.lang", langKey)
		languagePropertiesFilePath := filepath.Join(textsDir, languagePropertiesFileName)
		p, err := LoadLanguageProperties(languagePropertiesFilePath)
		if err != nil {
			// On error skip language properties
			continue
		}

		// Save these properties for this language
		if pack.LanguageMap == nil {
			pack.LanguageMap = make(map[string]*properties.Properties)
		}
		pack.LanguageMap[langKey] = p
	}

	return pack, nil
}

// LoadPackFromZip loads a pack stored in the given relative zip directory.
// The given relative directory path must contains a manifest.json file to be a valid pack.
// Returns a valid pack or an error otherwise.
func LoadPackFromZip(zipPath string, packDir string) (*Pack, error) {
	manifestRelativePath := ZipFilePathJoin(packDir, "manifest.json")

	// Get tje manifest json RAW bytes
	data, err := readZipEntry(zipPath, manifestRelativePath)
	if err != nil {
		return nil, err
	}

	// Parse it
	manifest, err := LoadManifestFromBytes(data)
	if err != nil {
		manifestAbsolutePath := ZipFilePathJoin(zipPath, packDir, "manifest.json")
		return nil, fmt.Errorf("failed to parse manifest in zip %q: %v", manifestAbsolutePath, err)
	}

	// Create the pack
	pack := &Pack{
		Path:     packDir,
		Manifest: manifest,
	}
	return pack, nil
}

// LoadPacksFromSubdirectories browse the sub directories from the given directory and loads a pack from each subdir.
// All sub directories must be a valid pack directory, otherwise the function returns an error.
// Returns a valid list of packs. Returns an empty list if there are no subdirectories.
// Returns an error otherwise.
func LoadPacksFromSubdirectories(path string) ([]*Pack, error) {
	var packs []*Pack

	// Get all the sub directories
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read multiple packs from directory %q: %v", path, err)
	}
	for _, e := range entries {
		if e.IsDir() {

			// Parse pack at this directory
			packDir := filepath.Join(path, e.Name())
			pack, err := LoadPackFromDirectory(packDir)
			if err != nil {
				return packs, err
			}

			packs = append(packs, pack)
		}
	}

	return packs, nil
}

// LoadAllPacksFromDirectoriesOrSubdirectories recursively traverses the given directory to detect and load directories containing Packs.
// This function is compatible with `.mcpack` and `.mcaddon` files.
// Extension `.mcpack`  have a manifest.json file at the root directory.
// Extension `.mcaddon` have a manifest.json file in each sub directory.
// Returns an error if a directory containing a manifest.json which fails to load as a pack.
// Returns am empty pack list when no manifest.json files is found.
func LoadAllPacksFromDirectoriesOrSubdirectories(root string) ([]*Pack, error) {
	var packs []*Pack

	err := filepath.WalkDir(root, func(path string, dir fs.DirEntry, err error) error {
		// If there is an error accessing a path, return it to stop walking
		if err != nil {
			return err
		}

		// Ignore files entirely
		if !dir.IsDir() {
			return nil
		}

		// Does directory contains a manifest.json file ?
		manifestPath := filepath.Join(path, "manifest.json")
		if fileExists(manifestPath) {
			// This directory contains a manifest, load this directory as a pack.
			pack, err := LoadPackFromDirectory(path)
			if err != nil {
				// Pack failed to load
				return err
			}

			packs = append(packs, pack)
		}
		return nil
	})

	if err != nil {
		// An error occured while walking the directories
		return nil, fmt.Errorf("failed to detect packs from directory %q: %v", root, err)
	}

	// Success
	return packs, nil
}

// LoadPacksFromZip recursively traverses the given zip file to detect and load directories containing Packs.
// This function is compatible with `.mcpack` and `.mcaddon` files.
// Extension `.mcpack`  have a manifest.json file at the root directory.
// Extension `.mcaddon` have a manifest.json file in each sub directory.
// Returns an error if a directory containing a manifest.json which fails to load as a pack.
// Returns am empty pack list when no manifest.json files is found.
func LoadPacksFromZip(zipPath string) ([]*Pack, error) {
	var packs []*Pack

	// find all manifests inside the addon
	relativeManifestPaths, err := FindManifestsRelativePathInAddon(zipPath)
	if err != nil {
		return nil, err
	}

	// for each manifest
	for _, relativeManifestPath := range relativeManifestPaths {
		relativePackDir := ZipFilePathGetParentDir(relativeManifestPath)
		if relativePackDir == "." {
			relativePackDir = "" // zip files do not support `.` and `..` directories
		}

		// Parse it
		pack, err := LoadPackFromZip(zipPath, relativePackDir)
		if err != nil {
			return nil, err
		}

		// Keep it
		packs = append(packs, pack)
	}

	// Success
	return packs, nil
}

// FindPackByUUID searches a given list of packs for a pack with the given UUID.
func FindPackByUUID(packs []*Pack, uuid string) *Pack {
	for _, pack := range packs {
		if strings.EqualFold(pack.Manifest.Header.UUID, uuid) {
			// This is the pack we are looking for
			return pack
		}
	}
	return nil
}

// FilterPacksByUUID filters a given list of packs by UUID.
// There should not be multiple packs with the same UUID in the same list.
// This function is mostly for cleanup and integrity.
func FilterPacksByUUID(packs []*Pack, uuid string) []*Pack {
	var results []*Pack
	for _, pack := range packs {
		if strings.EqualFold(pack.Manifest.Header.UUID, uuid) {
			// Match !
			results = append(results, pack)
		}
	}
	return results
}

// FilterPacksByKind filters a given list of packs by PackKind.
func FilterPacksByKind(packs []*Pack, kind PackKind) []*Pack {
	var results []*Pack
	for _, pack := range packs {
		if pack.KindSafe() == kind {
			// Match !
			results = append(results, pack)
		}
	}
	return results
}

func RemoveFormattingInPackName(name string) string {
	var b strings.Builder

	// for each rune
	skipNextRune := false
	for _, r := range name {
		if r == '§' {
			skipNextRune = true
		} else if skipNextRune {
			skipNextRune = false
		} else {
			b.WriteRune(r)
		}
	}

	result := b.String()
	return result
}

func LoadLanguageProperties(path string) (*properties.Properties, error) {
	p, err := properties.LoadFile(path, properties.UTF8)
	if err != nil {
		return nil, err
	}

	return p, err
}

func LoadLanguages(filePath string) ([]string, error) {
	bytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var languages []string
	err = json.Unmarshal(bytes, &languages)
	if err != nil {
		return nil, err
	}

	return languages, nil
}
