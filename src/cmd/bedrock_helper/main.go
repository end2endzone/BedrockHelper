// Command bedrock_helper is a command line tool for installing, uninstalling
// and inspecting Minecraft Bedrock Edition add-on packs (.mcaddon /
// .mcpack) on a Minecraft Bedrock Dedicated Server.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"runtime"
)

const usageText = `bedrock_helper - install and manage Minecraft Bedrock add-on packs.

Usage:
  bedrock_helper --install <path> [--server-location <dir>]
  bedrock_helper --uninstall <path-or-uuid> [--server-location <dir>]
  bedrock_helper --find-addons
  bedrock_helper --list-addons [--server-location <dir>]
  bedrock_helper --resolve-pack <uuid> [--server-location <dir>]
  bedrock_helper --install-all [--server-location <dir>]
  bedrock_helper --uninstall-all [--server-location <dir>]
  bedrock_helper --help

Flags:
  --install <path>          Install the .mcaddon/.mcpack/.zip add-on at <path>.
  --uninstall <path|uuid>   Uninstall the add-on at <path>, or by pack UUID
                             if the original add-on file is unavailable.
  --find-addons              Search the current directory (recursively) for
                             files that look like add-on packs and list them.
  --list-addons              List the add-on packs currently registered for
                             the target server, resolving each UUID to a name.
  --resolve-pack <uuid>      Search the target server for an add-on file that
                             contains a pack matching <uuid>.
  --install-all               Scan the target server directory for add-on
                             files and install every one that is found.
  --uninstall-all             Scan the target server directory for add-on
                             files and uninstall every one that is found.
  --server-location <dir>   Target Minecraft Bedrock server directory.
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
	fs := flag.NewFlagSet("bedrock_helper", flag.ContinueOnError)
	fs.SetOutput(os.Stdout)
	fs.Usage = func() { fmt.Fprint(os.Stdout, usageText) }

	var (
		installPath    string
		uninstallArg   string
		findAddons     bool
		listAddons     bool
		resolvePack    string
		installAll     bool
		uninstallAll   bool
		serverLocation string
	)

	fs.StringVar(&installPath, "install", "", "install an add-on pack file")
	fs.StringVar(&uninstallArg, "uninstall", "", "uninstall an add-on pack file or pack UUID")
	fs.BoolVar(&findAddons, "find-addons", false, "find add-on files in the current directory")
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
	for _, set := range []bool{installPath != "", uninstallArg != "", findAddons, listAddons, resolvePack != "", installAll, uninstallAll} {
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
	case findAddons:
		err = cmdFindAddons()
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

// ErrNotImplemented serves as our foundational sentinel error
var ErrNotImplemented = errors.New("not implemented")

// NotImplementedErr dynamically grabs the name of whichever function invokes it
func NotImplementedErr() error {
	// Get program counters of function invocations on the calling goroutine's stack.
	// Note: using 1 as argument to skip 'this function' and look at its immediately previous caller
	pc, _, _, ok := runtime.Caller(1)
	if !ok {
		return ErrNotImplemented
	}

	// Extracts the fully qualified function name (e.g., "main.Run")
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return ErrNotImplemented
	}

	return fmt.Errorf("function %s: %w", fn.Name(), ErrNotImplemented)
}

func cmdInstall(addonPath, serverLocation string) error {
	return NotImplementedErr()
}

func cmdUninstall(arg, serverLocation string) error {
	return NotImplementedErr()
}

func cmdFindAddons() error {
	return NotImplementedErr()
}

func cmdListAddons(serverLocation string) error {
	return NotImplementedErr()
}

func cmdResolvePack(uuid, serverLocation string) error {
	return NotImplementedErr()
}

func cmdInstallAll(serverLocation string) error {
	return NotImplementedErr()
}

func cmdUninstallAll(serverLocation string) error {
	return NotImplementedErr()
}
