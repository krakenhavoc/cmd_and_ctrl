// Command mcpseat is ADR 0122's local MCP seat: an MCP client (Claude
// Code, Codex) launches it over stdio, and it sits that client in a guest
// seat at a cmd_and_ctrl table, marked to the table as an AI agent.
//
// stdout carries the MCP protocol and nothing else; every log line goes to
// stderr, or to --log-file. See internal/mcpseat for the tools.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"strings"
	"syscall"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/mcpseat"
)

// version and commit are stamped by the release workflow
// (.github/workflows/mcpseat-release.yml) with
// -ldflags "-X main.version=… -X main.commit=…". A local build keeps these
// defaults, filled in from the Go build info where it has them.
var (
	version = "dev"
	commit  = "unknown"
)

type listFlag []string

func (l *listFlag) String() string     { return strings.Join(*l, ",") }
func (l *listFlag) Set(v string) error { *l = append(*l, v); return nil }

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "mcpseat:", err)
		os.Exit(1)
	}
}

// buildIdentity is the version and commit this binary reports. Stamped
// values win. Otherwise `go install …@vX` supplies the module version, and
// a build from a git checkout supplies the revision.
func buildIdentity(version, commit string, bi *debug.BuildInfo) (string, string) {
	if bi == nil {
		return version, commit
	}
	if version == "dev" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
		version = bi.Main.Version
	}
	if commit == "unknown" {
		for _, s := range bi.Settings {
			if s.Key == "vcs.revision" && s.Value != "" {
				commit = s.Value
			}
		}
	}
	return version, commit
}

// versionLine is what --version prints, for a bug report to quote.
func versionLine(version, commit string) string {
	return fmt.Sprintf("mcpseat %s (commit %s, %s %s/%s)", version, commit, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

func run(args []string, stdout io.Writer) error {
	bi, _ := debug.ReadBuildInfo()
	ver, rev := buildIdentity(version, commit, bi)

	fs := flag.NewFlagSet("mcpseat", flag.ExitOnError)
	var origins listFlag
	fs.Var(&origins, "allow-origin", "a server join may use, as scheme://host[:port] (port may be *); repeatable. Default: "+strings.Join(mcpseat.DefaultOrigins, ", "))
	absorb := fs.String("absorb", strings.Join(mcpseat.DefaultAbsorb, ","), "Layer A rules answered without the model: a subset of forced,mana-only,coin-call,same-land, or none")
	logFile := fs.String("log-file", "", "write the log here (mode 0600) instead of stderr")
	stateDir := fs.String("state-dir", "", "where saved sessions live (default $XDG_STATE_HOME/cmdctrl-mcpseat)")
	verbose := fs.Bool("v", false, "debug logging")
	showVersion := fs.Bool("version", false, "print the version and commit, and exit")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *showVersion {
		_, err := fmt.Fprintln(stdout, versionLine(ver, rev))
		return err
	}

	al, err := mcpseat.NewOriginAllowlist(origins)
	if err != nil {
		return err
	}
	rules, err := mcpseat.ParseAbsorb(*absorb)
	if err != nil {
		return err
	}
	var logOut io.Writer = os.Stderr
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600) // #nosec G304 -- the owner's own flag
		if err != nil {
			return err
		}
		defer func() { _ = f.Close() }()
		if err := f.Chmod(0o600); err != nil {
			return err
		}
		logOut = f
	}
	level := slog.LevelInfo
	if *verbose {
		level = slog.LevelDebug
	}
	seat, err := mcpseat.NewSeat(mcpseat.Config{
		Origins:   al,
		Absorb:    rules,
		StateDir:  *stateDir,
		LogOutput: logOut,
		LogLevel:  level,
	})
	if err != nil {
		return err
	}
	defer seat.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	seat.Logger().Info("mcpseat starting", "version", ver, "commit", rev, "origins", al.String(), "absorb", *absorb)
	if err := mcpseat.Serve(ctx, seat, ver); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}
