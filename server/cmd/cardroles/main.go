// Command cardroles writes data/scryfall/card-roles.json: each card's
// roles (removal, ramp, tutor, ...) for the advanced stats (ADR 0217 §7).
//
//	cardroles [-data-dir data] [-out path] [-in oracle-tags.jsonl.gz]
//
// The roles come from Scryfall's community Oracle tags (Tagger). It reads
// Scryfall's Oracle Tags bulk file, two requests a run, and maps tags to
// roles with the table in roles.go. It runs weekly after
// scripts/scryfall-refresh.sh (docs/commands.md). On any failure it
// writes nothing, so the last good file stays and the server keeps using
// it. -in reads a local copy of the bulk file instead of fetching.
package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

func main() {
	dataDir := flag.String("data-dir", envOr("CMDCTRL_DATA_DIR", "./data"),
		"server data dir (default $CMDCTRL_DATA_DIR, else ./data)")
	out := flag.String("out", "", "output file (default <data-dir>/scryfall/card-roles.json)")
	in := flag.String("in", "", "read this local Oracle Tags file (JSON Lines, gzip or plain) instead of fetching it")
	catalog := flag.String("catalog-url", defaultCatalogURL, "Scryfall bulk-data catalog URL")
	flag.Parse()
	if *out == "" {
		*out = filepath.Join(*dataDir, "scryfall", "card-roles.json")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	f := &fetcher{
		client:   &http.Client{Timeout: 5 * time.Minute},
		pause:    time.Second,
		throttle: 35 * time.Second,
		sleep:    time.Sleep,
	}
	built, err := run(ctx, f, *catalog, *in, *out, time.Now())
	if err != nil {
		fmt.Fprintln(os.Stderr, "cardroles:", err)
		os.Exit(1)
	}
	fmt.Printf("cardroles: wrote %s (%d cards with a role, from %s)\n", *out, len(built.Roles), built.Source)
	for _, rt := range roleTags {
		for _, slug := range rt.Tags {
			fmt.Printf("  %-17s <- %-28s %5d cards\n", rt.Role, slug, built.Tags[slug])
		}
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
