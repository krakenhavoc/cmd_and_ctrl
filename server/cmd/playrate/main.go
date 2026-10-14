// Command playrate prints catalog coverage by play rate (#3352): how many
// of the top-N EDHREC-ranked cards the effect catalog holds, and which
// top-N cards it is missing.
//
//	playrate [-dump data/scryfall/default-cards.json] [-top 500,1000,2000] [-list 1000]
//
// The dump defaults to $CMDCTRL_SCRYFALL_DUMP. -list is the rank cut-off
// for the missing-card list (0 prints the table only).
package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/coverage"
)

func main() {
	dump := flag.String("dump", os.Getenv("CMDCTRL_SCRYFALL_DUMP"),
		"Scryfall default-cards dump (default $CMDCTRL_SCRYFALL_DUMP)")
	top := flag.String("top", "", "comma-separated rank cut-offs (default 250,500,1000,2000,5000)")
	list := flag.Int("list", 1000, "list the missing cards at or under this rank (0 for none)")
	flag.Parse()
	if *dump == "" {
		fmt.Fprintln(os.Stderr, "playrate: -dump (or CMDCTRL_SCRYFALL_DUMP) is required")
		os.Exit(2)
	}
	tops := coverage.DefaultPlayRateTops
	if *top != "" {
		tops = nil
		for _, p := range strings.Split(*top, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || n <= 0 {
				fmt.Fprintf(os.Stderr, "playrate: bad -top value %q\n", p)
				os.Exit(2)
			}
			tops = append(tops, n)
		}
	}
	f, err := os.Open(*dump)
	if err != nil {
		fmt.Fprintln(os.Stderr, "playrate:", err)
		os.Exit(1)
	}
	defer func() { _ = f.Close() }()
	ranked, err := coverage.ReadRankedCards(f)
	if err != nil {
		fmt.Fprintln(os.Stderr, "playrate:", err)
		os.Exit(1)
	}
	fmt.Print(coverage.BuildPlayRateReport(ranked, coverage.CoveredOracleIDs(), tops, *list))
}
