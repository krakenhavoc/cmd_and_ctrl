package main

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func ptr(v float64) *float64 { return &v }

// The fixtures are four games and five picks of two drafts, written by
// hand so every number below can be counted from them:
//
//   - Alpha Bear is in hand in games 1 (opening hand, won), 2 (drawn,
//     lost) and 4 (drawn, won), and only in the deck in game 3: 3 games,
//     2 won. Beta, the Fox is in hand in games 2 (lost) and 3 (won); in
//     game 1 it was only tutored, which 17Lands does not count: 2 games,
//     1 won.
//   - Alpha Bear is taken at picks 9 and 3 (ATA 6). It is seen in draft
//     1's first pack at pick 1 and is still there when the pack comes
//     back at pick 9, so it was last seen at 9; in the second pack at
//     pick 3, which has no pick 11 row, so at 3 (ALSA 6). Beta, the Fox
//     is taken at pick 1 twice (ATA 1, ALSA 1).
//   - Plains has numbers like any card, and is left out.
//   - The rows of draft 1 are out of pick order, and draft 2 starts on
//     2026-01-02, the earliest day in either file.
var want = []Row{
	{Name: "Alpha Bear", EverDrawnGameCount: 3, EverDrawnWinRate: ptr(0.666667), AvgSeen: ptr(6), SeenCount: 2, AvgPick: ptr(6), PickCount: 2},
	{Name: "Beta, the Fox", EverDrawnGameCount: 2, EverDrawnWinRate: ptr(0.5), AvgSeen: ptr(1), SeenCount: 2, AvgPick: ptr(1), PickCount: 2},
}

func readFixtures(t *testing.T) *Stats {
	t.Helper()
	s := NewStats()
	if err := readFile("testdata/game_data.csv", s.ReadGames); err != nil {
		t.Fatal(err)
	}
	if err := readFile("testdata/draft_data.csv", s.ReadDrafts); err != nil {
		t.Fatal(err)
	}
	return s
}

func TestTheFixturesGiveTheHandCountedNumbers(t *testing.T) {
	s := readFixtures(t)
	got := s.Rows()
	if !reflect.DeepEqual(got, want) {
		g, _ := json.Marshal(got)
		w, _ := json.Marshal(want)
		t.Errorf("rows\n got %s\nwant %s", g, w)
	}
	if s.First() != "2026-01-02" {
		t.Errorf("first day %q, want 2026-01-02", s.First())
	}
	if s.games != 4 || s.picks != 5 {
		t.Errorf("read %d games and %d picks, want 4 and 5", s.games, s.picks)
	}
}

func TestACardWithNoGamesHasNullRates(t *testing.T) {
	s := NewStats()
	if err := s.ReadDrafts(strings.NewReader("draft_id,draft_time,pack_number,pick_number,pick,pack_card_Gamma\nd,2026-01-01,0,0,Gamma,1\n")); err != nil {
		t.Fatal(err)
	}
	rows := s.Rows()
	if len(rows) != 1 || rows[0].EverDrawnWinRate != nil || rows[0].EverDrawnGameCount != 0 || rows[0].AvgPick == nil {
		t.Errorf("rows = %+v", rows)
	}
	var b bytes.Buffer
	if err := WriteRows(&b, rows); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(b.String(), `"ever_drawn_win_rate": null`) {
		t.Errorf("a missing rate is not null: %s", b.String())
	}
}

func TestAMalformedFileIsAnError(t *testing.T) {
	for name, read := range map[string]func(*Stats) error{
		"games without won": func(s *Stats) error {
			return s.ReadGames(strings.NewReader("draft_time,opening_hand_A,drawn_A\n2026-01-01,1,0\n"))
		},
		"won is not a truth": func(s *Stats) error {
			return s.ReadGames(strings.NewReader("draft_time,won,opening_hand_A,drawn_A\n2026-01-01,maybe,1,0\n"))
		},
		"opening hand without drawn": func(s *Stats) error {
			return s.ReadGames(strings.NewReader("draft_time,won,opening_hand_A\n2026-01-01,True,1\n"))
		},
		"drafts without pick": func(s *Stats) error {
			return s.ReadDrafts(strings.NewReader("draft_id,draft_time,pack_number,pick_number\nd,2026-01-01,0,0\n"))
		},
		"pick number not a number": func(s *Stats) error {
			return s.ReadDrafts(strings.NewReader("draft_id,draft_time,pack_number,pick_number,pick\nd,2026-01-01,0,x,A\n"))
		},
		"short row": func(s *Stats) error {
			return s.ReadGames(strings.NewReader("draft_time,won,opening_hand_A,drawn_A\n2026-01-01,True\n"))
		},
	} {
		if err := read(NewStats()); err == nil {
			t.Errorf("%s: read without an error", name)
		}
	}
}

// run reads gzip as 17Lands publishes it, and plain CSV, and writes
// the rows the script builds a ratings file from.
func TestRunReadsGzipAndWritesTheRows(t *testing.T) {
	dir := t.TempDir()
	raw, err := os.ReadFile("testdata/game_data.csv")
	if err != nil {
		t.Fatal(err)
	}
	gz := filepath.Join(dir, "game_data_public.TST.PremierDraft.csv.gz")
	var b bytes.Buffer
	zw := gzip.NewWriter(&b)
	if _, err := zw.Write(raw); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(gz, b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(dir, "rows.json")
	if err := run(gz, "testdata/draft_data.csv", "", out); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	var got []Row
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("rows %s", data)
	}
	if _, err := os.Stat(out + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("the temporary file is left behind: %v", err)
	}
	if err := run(filepath.Join(dir, "missing.csv"), "testdata/draft_data.csv", "", filepath.Join(dir, "none.json")); err == nil {
		t.Error("a missing file is not an error")
	}
}
