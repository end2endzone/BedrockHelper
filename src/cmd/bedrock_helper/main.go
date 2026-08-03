package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	lib "bedrock_helper/libbedrockpacks"
)

const usageText = `Usage:
  bedrock_helper --install <path> [--server-location <dir>]
  bedrock_helper --uninstall <path-or-uuid> [--server-location <dir>]
  bedrock_helper --find-addons <path>
  bedrock_helper --list-addons [--server-location <dir>]
  bedrock_helper --resolve-pack <uuid> [--server-location <dir>]
  bedrock_helper --install-all [--server-location <dir>]
  bedrock_helper --uninstall-all [--server-location <dir>]
  bedrock_helper --help

Flags:
  --install <path>           Install the .mcaddon/.mcpack/.zip add-on at <path>.
  --uninstall <path|uuid>    Uninstall the add-on at <path>, or by pack UUID
                             if the original add-on file is unavailable.
  --find-addons <path>       Search the directory at <path> recursively for
  							 files that look like add-on packs and list them.
  --list-addons              List the add-on packs currently registered for
                             the target server.
  --resolve-pack <uuid>      Search the target server for an add-on file that
                             contains a pack matching <uuid>.
  --install-all              Scan the target server directory for add-on
                             files and install every one that is found.
  --uninstall-all            Scan the target server directory for add-on
                             files and uninstall every one that is found.
  --server-location <dir>    Target Minecraft Bedrock server directory.
                             Optional; defaults to the current directory.
  --help                     Show this usage message.

Examples:
  bedrock_helper --install $HOME/foobar.mcaddon --server-location $HOME/myserverinstalldir
  bedrock_helper --uninstall $HOME/foobar.mcaddon --server-location $HOME/myserverinstalldir
  bedrock_helper --uninstall 2bda6085-9d71-4d8a-9b9f-74e07b30459c --server-location $HOME/myserverinstalldir
  bedrock_helper --find-addons
  bedrock_helper --list-addons --server-location $HOME/myserverinstalldir
  bedrock_helper --resolve-pack "2bda6085-9d71-4d8a-9b9f-74e07b30459c" --server-location $HOME/myserverinstalldir
  bedrock_helper --install-all --server-location $HOME/myserverinstalldir
  bedrock_helper --uninstall-all --server-location $HOME/myserverinstalldir
`

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	fmt.Fprintf(os.Stderr, "bedrock_helper - install and manage Minecraft Bedrock add-on packs.\n")
	fmt.Fprintf(os.Stderr, "Version %s.\n\n", GetProductVersion())

	fs := flag.NewFlagSet("bedrock_helper", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() { fmt.Fprint(os.Stdout, usageText) }

	var (
		installPath    string
		uninstallArg   string
		findAddons     string
		listAddons     bool
		resolvePack    string
		installAll     bool
		uninstallAll   bool
		serverLocation string
	)

	fs.StringVar(&installPath, "install", "", "install an add-on pack file")
	fs.StringVar(&uninstallArg, "uninstall", "", "uninstall an add-on pack file or pack UUID")
	fs.StringVar(&findAddons, "find-addons", "", "find add-on files in the given directory or current directory")
	fs.BoolVar(&listAddons, "list-addons", false, "list add-ons registered for the server")
	fs.StringVar(&resolvePack, "resolve-pack", "", "resolve a pack UUID to an add-on file")
	fs.BoolVar(&installAll, "install-all", false, "install every add-on found on the server")
	fs.BoolVar(&uninstallAll, "uninstall-all", false, "uninstall every add-on found on the server")
	fs.StringVar(&serverLocation, "server-location", "", "target Minecraft Bedrock server directory")

	var err error

	// Parse arguments
	err = fs.Parse(args)
	if err != nil {
		if err == flag.ErrHelp {
			fs.Usage()
			return 0
		}
		fmt.Fprintf(os.Stderr, "error: invalid arguments: %v\n", err)
		return 2
	}

	// Check optional argument, set default value if unspecified
	if serverLocation == "" {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: could not determine current directory: %v\n", err)
			return 1
		}
		serverLocation = cwd
	}

	// Count how many commands are specified in the arguments
	commandsSet := 0
	for _, set := range []bool{installPath != "", uninstallArg != "", findAddons != "", listAddons, resolvePack != "", installAll, uninstallAll} {
		if set {
			commandsSet++
		}
	}

	// Act accordingly if too none or too many are specified
	switch {
	case commandsSet == 0:
		fs.Usage()
		return 0
	case commandsSet > 1:
		fmt.Fprintln(os.Stderr, "error: please specify only one command at a time")
		return 2
	}

	// Call the actual command helpers
	switch {
	case installPath != "":
		err = cmdInstall(installPath, serverLocation)
	case uninstallArg != "":
		err = cmdUninstall(uninstallArg, serverLocation)
	case findAddons != "":
		err = cmdFindAddons(findAddons)
	case listAddons:
		err = cmdListAddons(serverLocation)
	case resolvePack != "":
		err = cmdResolvePack(resolvePack, serverLocation)
	case installAll:
		err = cmdInstallAll(serverLocation)
	case uninstallAll:
		err = cmdUninstallAll(serverLocation)
	}

	// Check for an error while running a command.
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	return 0
}

// cmdInstall installs the .mcaddon/.mcpack/.zip add-on at <serverLocation>.
func cmdInstall(addonPath string, serverLocation string) error {
	installedPacks, err := lib.InstallAddonInServer(addonPath, serverLocation)
	if err != nil {
		return err
	}
	for _, p := range installedPacks {
		fmt.Printf("Installed %s\n", p.Description())
	}
	return nil
}

// cmdUninstall uninstalls the add-on at <serverLocation>, or by pack UUID if the original add-on file is unavailable.
func cmdUninstall(arg string, serverLocation string) error {
	// Check if arg is an addon file path or a UUID.

	// Check if arg is a file path and exists
	info, statErr := os.Stat(arg)
	if statErr == nil && !info.IsDir() {
		uninstalledPacks, err := lib.UninstallAddonInServer(arg, serverLocation)
		if err != nil {
			return err
		}
		for _, pack := range uninstalledPacks {
			fmt.Printf("Uninstalled %s\n", pack.Description())
		}
		return nil
	}

	// otherwise treat it as a pack UUID.
	pack, err := lib.UninstallPackInServerByUUID(arg, serverLocation)
	if err != nil {
		return err
	}
	fmt.Printf("Uninstalled %s\n", pack.Description())
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
	for _, a := range addons {
		fmt.Println("Found the following addons files:")
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

// cmdResolvePack search the target server for an add-on file that contains a pack matching <uuid>.
func cmdResolvePack(uuid, serverLocation string) error {
	path, err := lib.ResolvePackByUUID(uuid, serverLocation)
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
	for _, addonPath := range addons {
		err = cmdInstall(addonPath, serverLocation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error installing %s: %v\n", filepath.Base(addonPath), err)
			return err
		}
	}

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
	for _, addonPath := range addons {
		err = cmdUninstall(addonPath, serverLocation)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error uninstalling %s: %v\n", filepath.Base(addonPath), err)
			return err
		}
	}

	return nil
}
