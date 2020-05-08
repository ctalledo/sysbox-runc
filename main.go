package main

import (
	"fmt"
	"io"
	"os"

	"github.com/opencontainers/runtime-spec/specs-go"

	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

// Globals to be populated at build time during Makefile processing.
var (
	version  string // extracted from VERSION file
	commitId string // latest git commit-id of sysbox superproject
	builtAt  string // build time
	builtBy  string // build owner
)

const (
	specConfig = "config.json"
	usage      = `sysbox-runc

Nestybox's system container runtime.

Info: https://github.com/nestybox/sysbox-external
`
)

func main() {
	app := cli.NewApp()
	app.Name = "sysbox-runc"
	app.Usage = usage
	app.Version = version

	// show-version specialization.
	cli.VersionPrinter = func(c *cli.Context) {
		fmt.Printf("sysbox-runc\n"+
			"\tversion: \t%s\n"+
			"\tcommit: \t%s\n"+
			"\tbuilt at: \t%s\n"+
			"\tbuilt by: \t%s\n"+
			"\toci-specs: \t%s\n",
			c.App.Version, commitId, builtAt, builtBy, specs.Version)
	}

	root := "/run/sysbox-runc"
	if shouldHonorXDGRuntimeDir() {
		if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
			root = runtimeDir + "/sysbox-runc"
			// According to the XDG specification, we need to set anything in
			// XDG_RUNTIME_DIR to have a sticky bit if we don't want it to get
			// auto-pruned.
			if err := os.MkdirAll(root, 0700); err != nil {
				fatal(err)
			}
			if err := os.Chmod(root, 0700|os.ModeSticky); err != nil {
				fatal(err)
			}
		}
	}

	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "debug",
			Usage: "enable debug output for logging",
		},
		cli.StringFlag{
			Name:  "log",
			Value: "/dev/null",
			Usage: "set the log file path where internal debug information is written",
		},
		cli.StringFlag{
			Name:  "log-format",
			Value: "text",
			Usage: "set the format used by logs ('text' (default), or 'json')",
		},
		cli.StringFlag{
			Name:  "root",
			Value: root,
			Usage: "root directory for storage of container state (this should be located in tmpfs)",
		},
		cli.BoolFlag{
			Name:  "no-sysbox-fs",
			Usage: "do not interact with sysbox-fs; meant for testing and debugging.",
		},
		cli.BoolFlag{
			Name:  "no-sysbox-mgr",
			Usage: "do not interact with sysbox-mgr; meant for testing and debugging.",
		},
		cli.BoolFlag{
			Name:  "no-kernel-check",
			Usage: "do not check kernel compatibility; meant for testing and debugging.",
		},
		cli.BoolFlag{
			Name:   "cpu-profiling",
			Usage:  "enable cpu-profiling data collection; profile data is stored in the cwd of the process invoking sysbox-runc. Ignore the 'cannot set cpu profile rate' message (it's expected).",
			Hidden: true,
		},
		cli.BoolFlag{
			Name:   "memory-profiling",
			Usage:  "enable memory-profiling data collectionprofile data is stored in the cwd of the process invoking sysbox-runc.",
			Hidden: true,
		},
	}

	app.Commands = []cli.Command{
		createCommand,
		deleteCommand,
		eventsCommand,
		execCommand,
		initCommand,
		killCommand,
		listCommand,
		pauseCommand,
		psCommand,
		resumeCommand,
		runCommand,
		specCommand,
		startCommand,
		stateCommand,
		updateCommand,
	}

	app.Before = func(context *cli.Context) error {
		if context.GlobalBool("debug") {
			logrus.SetLevel(logrus.DebugLevel)
		}

		if path := context.GlobalString("log"); path != "" {
			f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_SYNC, 0666)
			if err != nil {
				return err
			}
			logrus.SetOutput(f)
		}

		switch context.GlobalString("log-format") {
		case "text":
			// retain logrus's default.
		case "json":
			logrus.SetFormatter(new(logrus.JSONFormatter))
		default:
			return fmt.Errorf("unknown log-format %q", context.GlobalString("log-format"))
		}

		return nil
	}

	// If the command returns an error, cli takes upon itself to print
	// the error on cli.ErrWriter and exit.
	// Use our own writer here to ensure the log gets sent to the right location.
	cli.ErrWriter = &FatalWriter{cli.ErrWriter}
	if err := app.Run(os.Args); err != nil {
		fatal(err)
	}
}

type FatalWriter struct {
	cliErrWriter io.Writer
}

func (f *FatalWriter) Write(p []byte) (n int, err error) {
	logrus.Error(string(p))
	return f.cliErrWriter.Write(p)
}
