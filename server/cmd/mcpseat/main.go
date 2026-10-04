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
	"strings"
	"syscall"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/mcpseat"
)

// version is stamped with -ldflags "-X main.version=…" when wanted.
var version = "dev"

type listFlag []string

func (l *listFlag) String() string     { return strings.Join(*l, ",") }
func (l *listFlag) Set(v string) error { *l = append(*l, v); return nil }

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "mcpseat:", err)
		os.Exit(1)
	}
}

func run() error {
	var origins listFlag
	flag.Var(&origins, "allow-origin", "a server join may use, as scheme://host[:port] (port may be *); repeatable. Default: "+strings.Join(mcpseat.DefaultOrigins, ", "))
	absorb := flag.String("absorb", strings.Join(mcpseat.DefaultAbsorb, ","), "Layer A rules answered without the model: a subset of forced,mana-only,coin-call,same-land,opening-roll,opening-choice, or none")
	logFile := flag.String("log-file", "", "write the log here (mode 0600) instead of stderr")
	stateDir := flag.String("state-dir", "", "where saved sessions live (default $XDG_STATE_HOME/cmdctrl-mcpseat)")
	verbose := flag.Bool("v", false, "debug logging")
	flag.Parse()

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
	seat.Logger().Info("mcpseat starting", "version", version, "origins", al.String(), "absorb", *absorb)
	if err := mcpseat.Serve(ctx, seat, version); err != nil && ctx.Err() == nil {
		return err
	}
	return nil
}
