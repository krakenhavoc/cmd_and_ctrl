// Command statsbackfill computes game statistics (ADR 0165 §6) for games
// that ended before the live hook existed, or under older definitions.
//
//	statsbackfill -data-dir /var/lib/cmd_and_ctrl/data [-dry-run]
//
// For each ended game with a replay (<data dir>/replays/<game id>.jsonl)
// it merges the frames' log windows by Seq, takes the last frame's view as
// the final view, and stores gamestats.Compute's record with source
// "replay". Games that already have a record of the current
// gamestats.Version are skipped; older records are computed again. It
// prints a summary and exits non-zero if any game failed.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/db"
)

func main() {
	dataDir := flag.String("data-dir", os.Getenv("CMDCTRL_DATA_DIR"),
		"server data directory (default $CMDCTRL_DATA_DIR)")
	dryRun := flag.Bool("dry-run", false, "compute and report, but write no records")
	flag.Parse()
	if *dataDir == "" {
		fmt.Fprintln(os.Stderr, "statsbackfill: -data-dir (or CMDCTRL_DATA_DIR) is required")
		os.Exit(2)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	d, err := db.Open(ctx, *dataDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "statsbackfill:", err)
		os.Exit(1)
	}
	defer func() { _ = d.Close() }()

	sum, err := run(ctx, d, *dataDir, *dryRun, time.Now(), os.Stdout)
	if err != nil {
		fmt.Fprintln(os.Stderr, "statsbackfill:", err)
		os.Exit(1)
	}
	if sum.Failed > 0 {
		os.Exit(1)
	}
}

// summary counts what a run did. Under -dry-run, Computed is what would
// have been written.
type summary struct {
	Computed int
	Skipped  int
	Failed   int
}

func (s summary) String() string {
	return fmt.Sprintf("computed %d, skipped %d, failed %d", s.Computed, s.Skipped, s.Failed)
}
