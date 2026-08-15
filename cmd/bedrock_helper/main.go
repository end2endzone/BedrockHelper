package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	lib "github.com/end2endzone/BedrockHelper/minecraftbedrock"
)

// Config holds all the command-line argument values
type Config struct {
	Install        string // path
	Uninstall      string // path
	FindAddons     string // path
	ResolvePack    string // uuid
	ServerLocation string // path
	ListAddons     bool
	InstallAll     bool
	UninstallAll   bool
	NoHeader       bool
	Version        bool
	Help           bool
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func printHeader() {
	fmt.Fprintf(os.Stdout, "bedrock_helper - install and manage Minecraft Bedrock add-on packs.\n")
}

func printVersion() {
	fmt.Fprintf(os.Stdout, "Version %s.\n", GetProductVersionString())
}

func reportArgumentParsingError(format string, args ...any) {
	// Prefix the message with a clear "Error:" label
	fullFormat := "Error: " + format + "\n"

	// Print the formatted string directly to standard error
	fmt.Fprintf(os.Stderr, fullFormat, args...)
}

// newBedrockFlagSet initializes the FlagSet and binds fields directly to the Config struct
func newBedrockFlagSet(cfg *Config) *flag.FlagSet {
	fs := flag.NewFlagSet("bedrock_helper", flag.ContinueOnError)

	// Bind string flags directly to the struct fields
	fs.StringVar(&cfg.Install, "install", "", "<path>|Install the .mcaddon/.mcpack/.zip add-on at <path>.")
	fs.StringVar(&cfg.Uninstall, "uninstall", "", "<path|uuid>|Uninstall the add-on at <path>, or by pack UUID if the original add-on file is unavailable.")
	fs.StringVar(&cfg.FindAddons, "find-addons", "", "<path>|Search the directory at <path> recursively for files that look like add-on packs and list them.")
	fs.StringVar(&cfg.ResolvePack, "resolve-pack", "", "<uuid>|Search the target server for an add-on file that contains a pack matching <uuid>.")
	fs.StringVar(&cfg.ServerLocation, "server-location", "", "<dir>|Target Minecraft Bedrock server directory.\nOptional; defaults to the current directory.")

	// Bind boolean flags directly to the struct fields
	fs.BoolVar(&cfg.ListAddons, "list-addons", false, "|List the add-on packs currently registered for the target server.")
	fs.BoolVar(&cfg.InstallAll, "install-all", false, "|Scan the target server directory for add-on files and install every one that is found.")
	fs.BoolVar(&cfg.UninstallAll, "uninstall-all", false, "|Scan the target server directory for add-on files and uninstall every one that is found.")
	fs.BoolVar(&cfg.NoHeader, "no-header", false, "|Do not show product header when running a command.")
	fs.BoolVar(&cfg.Version, "version", false, "|Show the product version.")

	// The flag library automatically registers `--help`, `-help` and `-h` flags.
	// We register the help flag specifically to make sure this flag is printed in our custom fs.Usage() function.
	fs.BoolVar(&cfg.Help, "help", false, "|Show this usage message.")

	// Attach the custom usage printer
	fs.Usage = func() {
		// This function can be called for multiple reasons:
		// 1. Parsing errors. When called, the parsing error is already printed to fs.Output().
		// 2. One of `--help`, `-help` or `-h` flags was used.
		//    Since we specifically register a manual `help` flag in our flagset, the flag library will not
		//    call this function for `--help` and `-help` but it does for `-h`.

		// Force usage text to be displayed on stdout by temporary switching output to stdout
		originalOutput := fs.Output() // os.Stderr or os.Stdout
		fs.SetOutput(os.Stdout)

		// Show an empty line between the error and the usage text (on stdout)
		fmt.Fprintln(os.Stdout)

		// Print usage text
		printUsage(fs)

		// Restore output to whatever it was set
		fs.SetOutput(originalOutput)
	}

	return fs
}

// getOrderedFlags visits all flags in a flagset and make a slice with them.
// The returned slices is also ordered so that `version` and `help` flags are last.
func getOrderedFlags(fs *flag.FlagSet) []*flag.Flag {
	flags := make([]*flag.Flag, 0)

	// Create a slice with all flags, skipping some flags...
	fs.VisitAll(func(f *flag.Flag) {
		if f.Name == "version" ||
			f.Name == "help" ||
			f.Name == "no-header" {
			return // skip
		}
		flags = append(flags, f)
	})

	// Add our bottom flags at the end of the list.
	flags = append(flags, fs.Lookup("no-header"))
	flags = append(flags, fs.Lookup("version"))
	flags = append(flags, fs.Lookup("help"))

	return flags
}

// StringSplitAtLast splits a given string at the last occurance of the given separator.
func StringSplitAtLast(s, separator string) []string {
	// Find the last occurrence of the separator
	i := strings.LastIndex(s, separator)

	// If the separator is not found, return the original string as a single-element slice
	if i == -1 {
		return []string{s}
	}

	// Slice the string before and after the last separator
	return []string{
		s[:i],
		s[i+len(separator):],
	}
}

// printUsage print a usage string that output each arguments and then examples.
func printUsage(fs *flag.FlagSet) {
	output := fs.Output() // os.Stderr

	// Print static usage header
	const usageText = `Usage:
    bedrock_helper --install <path> [--server-location <dir>] [--no-header]
    bedrock_helper --uninstall <path-or-uuid> [--server-location <dir>] [--no-header]
    bedrock_helper --find-addons <path> [--no-header]
    bedrock_helper --list-addons [--server-location <dir>] [--no-header]
    bedrock_helper --resolve-pack <uuid> [--server-location <dir>] [--no-header]
    bedrock_helper --install-all [--server-location <dir>] [--no-header]
    bedrock_helper --uninstall-all [--server-location <dir>] [--no-header]
    bedrock_helper --version
    bedrock_helper --help
	`
	fmt.Fprintln(output, usageText)

	// Print all flags and their descriptions in a 2 columns layout.
	// Column 0 is the name of the argument and its value descriptor such as `<path>`.
	// Column 1 is the flag's usage description. The description can contain \n character to force the following text to be displayed on the next line.
	// For example:
	// |----------------------------|---------------------------------------------------|
	// `  --find-addons <path>       Search the directory at <path> recursively for		`
	// `                             files that look like add-on packs and list them.	`

	fmt.Fprintln(output, "Flags:")

	orderedFlags := getOrderedFlags(fs)
	for _, f := range orderedFlags {
		// Split our custom usage metadata format: `<placeholder>|Description string`.
		// Using StringSplitAtLast() instead of strings.SplitN(f.Usage, "|", 2) to support
		// placeholders that contains optional names.
		// For example `<path|uuid>`.
		parts := StringSplitAtLast(f.Usage, "|")

		placeholder := ""
		description := f.Usage

		if len(parts) == 2 {
			placeholder = parts[0]
			description = parts[1]
		}

		// Construct flag component (for example "--install <path>")
		flagStr := "  --" + f.Name
		if placeholder != "" {
			flagStr += " " + placeholder
		}

		// Split the description into multiple lines to handle clean indentation alignment
		lines := strings.Split(description, "\n")

		// Target column where the 2nd column (description text) must begin
		const targetCol = 30

		// Print the first line
		if len(flagStr) < targetCol {
			// Pad the remaining space up to column targetCol
			padding := strings.Repeat(" ", targetCol-len(flagStr))
			fmt.Fprintf(output, "%s%s%s\n", flagStr, padding, lines[0])
		} else {
			// If the flag declaration is too long that it breaks our alignement...

			// Print it on its own line
			fmt.Fprintln(output, flagStr)

			// Then align the description
			padding := strings.Repeat(" ", targetCol)
			fmt.Fprintf(output, "%s%s\n", padding, lines[0])
		}

		// Print any multi-line description wrap-arounds exactly at column targetCol
		for i := 1; i < len(lines); i++ {
			padding := strings.Repeat(" ", targetCol)
			fmt.Fprintf(output, "%s%s\n", padding, lines[i])
		}
	}
	fmt.Fprintln(output)

	// Print static examples
	const exampleText = `Examples:
	bedrock_helper --install $HOME/foobar.mcaddon --server-location $HOME/myserverinstalldir
	bedrock_helper --uninstall $HOME/foobar.mcaddon --server-location $HOME/myserverinstalldir
	bedrock_helper --uninstall 2bda6085-9d71-4d8a-9b9f-74e07b30459c --server-location $HOME/myserverinstalldir
	bedrock_helper --find-addons /tmp/addons
	bedrock_helper --list-addons --server-location $HOME/myserverinstalldir
	bedrock_helper --resolve-pack \"2bda6085-9d71-4d8a-9b9f-74e07b30459c\" --server-location $HOME/myserverinstalldir
	bedrock_helper --install-all --server-location $HOME/myserverinstalldir
	bedrock_helper --uninstall-all --server-location $HOME/myserverinstalldir
	`
	fmt.Fprintln(output, exampleText)
}

func run(args []string) int {
	var cfg Config
	fs := newBedrockFlagSet(&cfg)

	// Manually parse for `--no-header` and `--version` arguments before calling fs.Parse().
	// In case of parsing errors, the error will be printed before the flag library will call our custom fs.Usage().
	// So we must print the application's header or version before doing the actual parsing.
	cfg.NoHeader = *flag.Bool("no-header", false, "") // hasArgument("--no-header")
	cfg.Version = *flag.Bool("version", false, "")    // hasArgument("--version")

	// Should we only print the version ?
	if cfg.Version {
		printVersion()
		return 0
	}

	// Print application header, unless specified not to
	if !cfg.NoHeader {
		printHeader()
		printVersion()
	}

	var err error

	// Parse arguments
	fs.SetOutput(os.Stderr) // parsing errors should be printed to stderr
	err = fs.Parse(args)
	fs.SetOutput(os.Stdout) // after parsing, following outputs should be printed to stdout
	if err != nil {

		// The flag library automatically registers `--help`, `-help` and `-h` flags.
		// When specified on the command line, these flags reports the specific error `flag.ErrHelp` on parsing.
		// Since we specifically register a manual `help` flag in our flagset, the flag library do not report
		// the error `flag.ErrHelp` for `--help` and `-help` but it does for `-h`.
		if err == flag.ErrHelp {
			// There is no need to call printUsage(fs) since the flag library has already called fs.Usage() because of the error.
			return 0
		}

		reportArgumentParsingError("invalid arguments: %v", err)
		return 2
	}

	// Help flag set
	if cfg.Help {
		// Triggered by `--help`, `-help`
		printUsage(fs)
		return 0
	}

	// Check optional argument, set default value if unspecified
	if cfg.ServerLocation == "" {
		cwd, err := os.Getwd()
		if err != nil {
			reportArgumentParsingError("could not determine current directory: %v", err)
			return 1
		}
		cfg.ServerLocation = cwd
	}

	// Count how many commands are specified in the arguments
	// Do not count `--version` and `--help` as these were already processed above.
	commandsSet := 0
	for _, set := range []bool{cfg.Install != "", cfg.Uninstall != "", cfg.FindAddons != "", cfg.ListAddons, cfg.ResolvePack != "", cfg.InstallAll, cfg.UninstallAll} {
		if set {
			commandsSet++
		}
	}

	// Act accordingly if too none or too many are specified
	switch {
	case commandsSet == 0:
		// No command specified. Do not know that to do.
		fmt.Fprint(os.Stderr, "no command specified\n\n")

		// Then show help message.
		printUsage(fs)
		return 2
	case commandsSet > 1:
		reportArgumentParsingError("please specify only one command at a time")
		return 2
	}

	// Call the actual command helpers
	switch {
	case cfg.Install != "":
		err = cmdInstall(cfg.Install, cfg.ServerLocation)
	case cfg.Uninstall != "":
		err = cmdUninstall(cfg.Uninstall, cfg.ServerLocation)
	case cfg.FindAddons != "":
		err = cmdFindAddons(cfg.FindAddons)
	case cfg.ListAddons:
		err = cmdListAddons(cfg.ServerLocation)
	case cfg.ResolvePack != "":
		err = cmdResolveAddon(cfg.ResolvePack, cfg.ServerLocation)
	case cfg.InstallAll:
		err = cmdInstallAll(cfg.ServerLocation)
	case cfg.UninstallAll:
		err = cmdUninstallAll(cfg.ServerLocation)
	}

	// Check for an error while running a command.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

func printCommandReport(command lib.Command, packs []*lib.Pack, serverLocation string) {
	fmt.Printf("%sed the following packs:\n", command.String())
	for _, p := range packs {
		fmt.Printf("  - %s\n", p.Description())
	}
	fmt.Printf("in server %v\n", serverLocation)
}

// cmdInstall installs the .mcaddon/.mcpack/.zip add-on at <serverLocation>.
func cmdInstall(addonPath string, serverLocation string) error {
	installedPacks, err := lib.InstallAddonInServer(addonPath, serverLocation)
	if err != nil {
		return err
	}

	// print report
	printCommandReport(lib.Install, installedPacks, serverLocation)

	return nil
}

// processUninstall uninstalls the add-on at <serverLocation>, or by pack UUID if the original add-on file is unavailable.
func processUninstall(arg string, serverLocation string) ([]*lib.Pack, error) {
	// Check if arg is an addon file path or a UUID.

	// Check if arg is a file path and exists
	info, statErr := os.Stat(arg)
	if statErr == nil && !info.IsDir() {
		uninstalledPacks, err := lib.UninstallAddonInServer(arg, serverLocation)
		if err != nil {
			return nil, err
		}

		return uninstalledPacks, nil
	}

	// otherwise treat it as a pack UUID.
	pack, err := lib.UninstallPackInServerByUUID(arg, serverLocation)
	if err != nil {
		return nil, err
	}

	return []*lib.Pack{pack}, nil
}

func cmdUninstall(arg string, serverLocation string) error {
	uninstalledPacks, err := processUninstall(arg, serverLocation)
	if err != nil {
		return err
	}

	// print report
	printCommandReport(lib.Uninstall, uninstalledPacks, serverLocation)

	return nil
}

// cmdFindAddons search the given directory recursively for files that look like add-on packs and list them.
func cmdFindAddons(findAddons string) error {
	addons, err := lib.FindAddonsInDir(findAddons, true)
	if err != nil {
		return err
	}

	if len(addons) == 0 {
		fmt.Println("No add-on files found.")
		return nil
	}

	// List hist
	fmt.Println("Found the following addons files:")
	for _, a := range addons {
		fmt.Printf("  * %v", a)
	}
	return nil
}

// cmdListAddons list the add-on packs currently registered for the target server, resolving each UUID to a name.
func cmdListAddons(serverLocation string) error {
	packs, err := lib.ListInstalledPacks(serverLocation)
	if err != nil {
		return err
	}
	if len(packs) == 0 {
		fmt.Println("No add-ons are registered for this server.")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "KIND\tNAME\tVERSION\tUUID")
	for _, p := range packs {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", p.KindSafe(), p.Name(), p.Manifest.Header.Version, p.Manifest.Header.UUID)
	}
	return w.Flush()
}

// cmdResolveAddon search the target server for an add-on file that contains a pack matching <uuid>.
func cmdResolveAddon(uuid, serverLocation string) error {
	path, err := lib.ResolveAddonByUUID(uuid, serverLocation)
	if err != nil {
		return err
	}

	// print only the path in the output so that it can be parsed by scripts
	fmt.Println(path)

	return nil
}

// cmdInstallAll scan the target server directory for add-on files and install every one that is found.
func cmdInstallAll(serverLocation string) error {
	addons, err := lib.FindAddonsInDir(serverLocation, true)
	if err != nil {
		return err
	}
	if len(addons) == 0 {
		fmt.Println("No add-on files found on the server.")
		return nil
	}

	// Process each identified addon as if they were specified individually in the command line
	installedPacks := make([]*lib.Pack, 0)
	for _, addonPath := range addons {

		// Install this addon
		latestInstalledPacks, err := lib.InstallAddonInServer(addonPath, serverLocation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error installing %s: %v\n", filepath.Base(addonPath), err)
			return err
		}

		// Append latest installation to total installation list
		installedPacks = append(installedPacks, latestInstalledPacks...)
	}

	// print report
	printCommandReport(lib.Install, installedPacks, serverLocation)

	return nil
}

func cmdUninstallAll(serverLocation string) error {
	addons, err := lib.FindAddonsInDir(serverLocation, true)
	if err != nil {
		return err
	}
	if len(addons) == 0 {
		fmt.Println("No add-on files found on the server.")
		return nil
	}

	// Process each identified addon as if they were specified individually in the command line
	uninstalledPacks := make([]*lib.Pack, 0)
	for _, addonPath := range addons {

		// Uninstall this addon
		latestUninstalledPacks, err := processUninstall(addonPath, serverLocation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error uninstalling %s: %v\n", filepath.Base(addonPath), err)
			return err
		}

		// Append latest uninstallation to total uninstallation list
		uninstalledPacks = append(uninstalledPacks, latestUninstalledPacks...)
	}

	// print report
	printCommandReport(lib.Uninstall, uninstalledPacks, serverLocation)

	return nil
}
