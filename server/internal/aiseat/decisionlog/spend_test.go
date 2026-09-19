package decisionlog_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/decisionlog"
)

// spend_test.go is #735's decision-log half: one spend record per game
// file, and every reader that was counting decision windows before it
// existed still counts exactly the same number.
//
// The second half is the one that would break silently. A JSONL file
// that grows a new line shape breaks the readers that do not know
// about it — `boteval`'s position harvest, the absorption census, the
// replay — and it breaks them by giving each a plausible record full
// of zeroes rather than by failing. Scan therefore filters and
// ScanAll is the way in.

func spendRecord() aiseat.GameSpend {
	return aiseat.GameSpend{
		Game: gameID,
		Seats: []aiseat.SeatSpend{
			{
				Seat: seatID, Policy: "assisted",
				Spend: aiseat.Spend{
					Decision:      aiseat.PurposeSpend{Calls: 31, Usage: aiseat.TokenUsage{InputTokens: 74000, OutputTokens: 380}},
					Improvisation: aiseat.PurposeSpend{Calls: 2, Usage: aiseat.TokenUsage{InputTokens: 2100, OutputTokens: 190}},
				},
			},
			{
				Seat: oppID, Policy: "heuristic",
			},
		},
	}
}

func TestTheGamesSpendRecordIsWrittenAndReadBack(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeAll, 0)
	for i := 0; i < 3; i++ {
		g.Observe(event(uint64(i+1), "C"))
	}
	want := spendRecord()
	g.ObserveSpend(want)
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	path := decisionlog.Path(dir, gameID)

	// Scan sees the windows and NOT the spend record: every reader
	// that counted decisions before #735 counts the same number now.
	windows := 0
	if err := decisionlog.Scan(path, func(r decisionlog.Record) error {
		windows++
		if r.Kind != "" {
			t.Errorf("Scan handed out a %q record; it is the decision-window walk", r.Kind)
		}
		if r.Spend != nil {
			t.Error("a decision window carries a spend record")
		}
		return nil
	}); err != nil {
		t.Fatalf("Scan: %v", err)
	}
	if windows != 3 {
		t.Errorf("Scan saw %d decision windows, want 3", windows)
	}

	// ScanAll sees both, and the spend record survived the round trip.
	var got []decisionlog.Record
	if err := decisionlog.ScanAll(path, func(r decisionlog.Record) error {
		got = append(got, r)
		return nil
	}); err != nil {
		t.Fatalf("ScanAll: %v", err)
	}
	if len(got) != 4 {
		t.Fatalf("ScanAll saw %d lines, want 4 (3 windows + 1 spend)", len(got))
	}
	last := got[len(got)-1]
	if last.Kind != decisionlog.KindSpend {
		t.Fatalf("the last line is %q, want %q — the spend record is written when the game is over",
			last.Kind, decisionlog.KindSpend)
	}
	if last.Spend == nil {
		t.Fatal("the spend record has no spend on it")
	}
	if last.Game != gameID.String() {
		t.Errorf("spend record names game %q", last.Game)
	}
	if len(last.Spend.Seats) != 2 {
		t.Fatalf("spend record has %d seats, want 2", len(last.Spend.Seats))
	}
	if last.Spend.Seats[0].Spend != want.Seats[0].Spend {
		t.Errorf("seat 0 round-tripped as %+v, want %+v", last.Spend.Seats[0].Spend, want.Seats[0].Spend)
	}
	if !last.Spend.Seats[1].Spend.Empty() {
		t.Errorf("the heuristic seat round-tripped as %+v, want nothing", last.Spend.Seats[1].Spend)
	}
	total := last.Spend.Total()
	if total.Decision.Calls != 31 || total.Improvisation.Calls != 2 {
		t.Errorf("round-tripped total = %d / %d calls, want 31 / 2",
			total.Decision.Calls, total.Improvisation.Calls)
	}
}

// A game that spent nothing still writes its line: "this table cost
// nothing" is a measurement, and a record that only appears when
// there is a bill cannot be told from one that was dropped.
func TestAFreeGameStillWritesItsSpendRecord(t *testing.T) {
	_, g, dir := newLog(t, decisionlog.ModeEscalated, 0)
	g.ObserveSpend(aiseat.GameSpend{
		Game:  gameID,
		Seats: []aiseat.SeatSpend{{Seat: uuid.New(), Policy: "heuristic"}},
	})
	if err := g.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	n := 0
	if err := decisionlog.ScanAll(decisionlog.Path(dir, gameID), func(r decisionlog.Record) error {
		if r.Kind == decisionlog.KindSpend {
			n++
			if r.Spend == nil || !r.Spend.Empty() {
				t.Errorf("a free game's record is %+v, want an empty spend", r.Spend)
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("ScanAll: %v", err)
	}
	if n != 1 {
		t.Errorf("a free game wrote %d spend records, want 1", n)
	}
}
