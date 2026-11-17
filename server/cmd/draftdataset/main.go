// Command draftdataset turns one set's 17Lands public datasets into card
// ratings, for a set whose live ratings endpoint has too little data
// (#5708). scripts/draft-ratings-refresh.sh --from-dataset runs it.
//
//	draftdataset -games game_data_public.BLB.PremierDraft.csv.gz \
//	             -drafts draft_data_public.BLB.PremierDraft.csv.gz \
//	             [-dump default-cards.json] -out rows.json
//
// 17Lands publishes, per set and format, every game and every pick its
// users logged (https://www.17lands.com/public_datasets). From them this
// computes, per card name, what the live endpoint serves: the games with
// the card in hand and their win rate (GIH WR), the average pick it was
// taken at (ATA) and the average pick it was last seen at (ALSA). The
// output is a JSON array in the live endpoint's shape, so the script's
// one builder applies the same threshold and writes the same file.
//
// It prints the day of the earliest draft read, "YYYY-MM-DD", on
// stdout: the file's "start", which the live import reads from 17Lands'
// filters. Both inputs may be gzip-compressed or plain CSV; they are
// read as a stream, never unpacked to disk.
//
// The rows keep 17Lands' spelling of a card, as the live import does;
// the server resolves it to a card at start (draftratings.Bind). With
// -dump (default $CMDCTRL_SCRYFALL_DUMP), a name the datasets mangled
// (a letter outside Latin-1 written "?") is put back from the dump.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	games := flag.String("games", "", "the set's game dataset (game_data_public.<SET>.<FORMAT>.csv[.gz])")
	drafts := flag.String("drafts", "", "the set's draft dataset (draft_data_public.<SET>.<FORMAT>.csv[.gz])")
	out := flag.String("out", "", "where to write the rows (JSON)")
	dump := flag.String("dump", os.Getenv("CMDCTRL_SCRYFALL_DUMP"),
		"Scryfall default-cards dump, to repair mangled names (default $CMDCTRL_SCRYFALL_DUMP; optional)")
	flag.Parse()
	if *games == "" || *drafts == "" || *out == "" {
		fmt.Fprintln(os.Stderr, "draftdataset: -games, -drafts and -out are required")
		os.Exit(2)
	}
	if err := run(*games, *drafts, *dump, *out); err != nil {
		fmt.Fprintln(os.Stderr, "draftdataset:", err)
		os.Exit(1)
	}
}

func run(games, drafts, dump, out string) error {
	s := NewStats()
	if err := readFile(games, s.ReadGames); err != nil {
		return fmt.Errorf("%s: %w", games, err)
	}
	if err := readFile(drafts, s.ReadDrafts); err != nil {
		return fmt.Errorf("%s: %w", drafts, err)
	}
	if s.games == 0 || s.picks == 0 {
		return fmt.Errorf("no rows: %d games, %d picks", s.games, s.picks)
	}
	if dump != "" {
		var known []string
		err := readFile(dump, func(r io.Reader) error {
			var err error
			known, err = ReadNames(r)
			return err
		})
		if err != nil {
			return fmt.Errorf("%s: %w", dump, err)
		}
		for _, line := range s.RepairNames(known) {
			fmt.Fprintln(os.Stderr, "draftdataset:", line)
		}
	}
	tmp := out + ".tmp"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if err := WriteRows(f, s.Rows()); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmp, out); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "draftdataset: %d games, %d picks, %d cards\n", s.games, s.picks, len(s.cards))
	fmt.Println(s.First())
	return nil
}

// readFile opens a dataset and hands it to read.
func readFile(path string, read func(io.Reader) error) error {
	r, err := Open(path)
	if err != nil {
		return err
	}
	defer func() { _ = r.Close() }()
	return read(r)
}
