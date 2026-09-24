// Command snapshotscrub removes player data from a restore-point file
// so it can be committed to the snapshot fixture corpus (#522).
//
//	go run ./cmd/snapshotscrub -in /tmp/restore/<id>.json \
//	    -out internal/game/testdata/snapshots/real/<name>.json
//
// Names, display names and player IDs become deterministic
// placeholders, Discord IDs and avatar hashes are removed, and the
// game ID is replaced by a hash of itself. The tool refuses to write
// anything if a Discord snowflake, an email address, or an original
// player name is still in the output; see internal/snapshotscrub.
// It never overwrites an existing file: corpus fixtures are
// append-only.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/snapshotscrub"
)

func main() {
	in := flag.String("in", "", "restore-point file to scrub (restore/<id>.json)")
	out := flag.String("out", "", "where to write the scrubbed file (must not exist)")
	allowNames := flag.Bool("allow-name-substrings", false,
		"warn, rather than refuse, when a string still contains an original player name — only after reading the file")
	flag.Parse()
	if *in == "" || *out == "" {
		flag.Usage()
		os.Exit(2)
	}
	if err := run(*in, *out, snapshotscrub.Options{AllowNameSubstrings: *allowNames}); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(in, out string, opt snapshotscrub.Options) error {
	if _, err := os.Stat(out); err == nil {
		return fmt.Errorf("%s already exists; corpus fixtures are never rewritten — pick a new name", out)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	raw, err := os.ReadFile(in)
	if err != nil {
		return err
	}
	clean, rep, err := snapshotscrub.Scrub(raw, opt)
	if err != nil {
		return err
	}
	for _, w := range rep.Warnings {
		fmt.Fprintln(os.Stderr, "warning:", w)
	}
	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(out, clean, 0o644); err != nil {
		return err
	}
	fmt.Printf("wrote %s: %d seats, %d strings replaced\n", out, rep.Seats, rep.Replaced)
	return nil
}
