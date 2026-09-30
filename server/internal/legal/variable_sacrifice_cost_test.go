package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// variable_sacrifice_cost_test.go — ADR 0100 §6: a variable sacrifice
// as a cast's additional cost is offered as zero (when the floor allows
// it) and then up to three positive counts, smallest first; each count
// is priced on its own through the one pricer, so a per-sacrifice
// discount is offered exactly where the engine charges it; and for
// "sacrifice X" the count IS the X. Every move is one the engine
// accepts (#544).

const (
	oracleViciousBetrayal         = "be1015a7-2ace-4b97-8884-202abae8401b"
	oracleTorgaar                 = "4229140f-fa5b-4727-a16a-cbbd756d979e"
	oracleEliminateTheCompetition = "26e549c0-a08b-475b-9138-6dde175cdf55"
	oracleDevastatingSummons      = "5eb6626a-1e0a-438e-9e95-a1e86be6489d"
)

type varSacPayload struct {
	XValue       int      `json:"x_value"`
	SacrificeIDs []string `json:"sacrifice_ids"`
	Targets      []struct {
		ID string `json:"id"`
	} `json:"targets"`
}

func varSacMovesOf(t *testing.T, moves []legal.Move, src uuid.UUID) []varSacPayload {
	t.Helper()
	var out []varSacPayload
	for _, m := range moves {
		if m.Source != src || m.Kind != legal.KindCast {
			continue
		}
		var p varSacPayload
		if err := json.Unmarshal(m.Params, &p); err != nil {
			t.Fatalf("decode %s: %v", m.Params, err)
		}
		out = append(out, p)
	}
	return out
}

func sacCounts(ps []varSacPayload) map[int]bool {
	out := map[int]bool{}
	for _, p := range ps {
		out[len(p.SacrificeIDs)] = true
	}
	return out
}

// Vicious Betrayal with four creatures: zero, one, two and three are
// offered — never four, the ladder's cap — and each is dispatchable.
func TestVariableSacrificeOffersZeroAndThreeCounts(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 5, "Swamp")
	for i := 0; i < 4; i++ {
		battlefieldCard(g, seat, creature("Bear", "{2}", 2, 2))
	}
	vb := handCard(seat, game.Card{Name: "Vicious Betrayal", TypeLine: "Sorcery", ManaCost: "{3}{B}{B}", OracleID: oracleViciousBetrayal})
	moves := legal.EnumerateFor(g, seat.ID)
	got := sacCounts(varSacMovesOf(t, moves, vb))
	for _, n := range []int{0, 1, 2, 3} {
		if !got[n] {
			t.Errorf("count %d not offered (offered %v)", n, got)
		}
	}
	if got[4] {
		t.Errorf("count 4 offered past the ladder's cap: %v", got)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Torgaar with two Swamps and three creatures: {6}{B}{B} is out of
// reach, and only the three-creature payment brings it to {B}{B}. The
// enumerator prices each count, so that is the one move offered.
func TestVariableSacrificePricesTheDiscount(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 2, "Swamp")
	for i := 0; i < 3; i++ {
		battlefieldCard(g, seat, creature("Bear", "{2}", 2, 2))
	}
	tg := handCard(seat, game.Card{Name: "Torgaar, Famine Incarnate", TypeLine: "Legendary Creature — Avatar",
		ManaCost: "{6}{B}{B}", OracleID: oracleTorgaar, Power: 7, Toughness: 6})
	moves := legal.EnumerateFor(g, seat.ID)
	got := varSacMovesOf(t, moves, tg)
	if len(got) != 1 || len(got[0].SacrificeIDs) != 3 {
		t.Fatalf("Torgaar moves = %+v, want exactly the three-creature payment", got)
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Eliminate the Competition: the targets fix X, and X fixes the one
// payment — k targets, k creatures sacrificed, x_value k. Never more
// targets than the seat has creatures to sacrifice, and never X = 0.
func TestSacrificeXFollowsTheTargets(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 5, "Swamp")
	for i := 0; i < 2; i++ {
		battlefieldCard(g, seat, creature("Bear", "{2}", 2, 2))
	}
	for i := 0; i < 3; i++ {
		battlefieldCard(g, g.Seats[1], creature("Ogre", "{2}", 3, 3))
	}
	etc := handCard(seat, game.Card{Name: "Eliminate the Competition", TypeLine: "Sorcery", ManaCost: "{4}{B}", OracleID: oracleEliminateTheCompetition})
	moves := legal.EnumerateFor(g, seat.ID)
	got := varSacMovesOf(t, moves, etc)
	if len(got) == 0 {
		t.Fatal("Eliminate the Competition not offered")
	}
	for _, p := range got {
		k := len(p.Targets)
		if k < 1 || k > 2 || len(p.SacrificeIDs) != k || p.XValue != k {
			t.Errorf("move = %+v: want 1–2 targets, as many sacrifices, and X equal to both", p)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}

// Devastating Summons: no targets, so the X is the sacrifice count on
// the ladder, starting at one (X = 0 makes two 0/0s, #810). The Mountain
// that pays {R} is never also sacrificed (#1242).
func TestSacrificeXLadderWithoutTargets(t *testing.T) {
	g := newTable(t)
	seat := g.Seats[g.Turn.ActiveSeat]
	clearHand(seat)
	advanceTo(t, g, game.StepPrecombatMain)
	basicLands(g, seat, 3, "Mountain")
	ds := handCard(seat, game.Card{Name: "Devastating Summons", TypeLine: "Sorcery", ManaCost: "{R}", OracleID: oracleDevastatingSummons})
	moves := legal.EnumerateFor(g, seat.ID)
	got := varSacMovesOf(t, moves, ds)
	counts := sacCounts(got)
	if counts[0] || !counts[1] || !counts[2] || counts[3] {
		t.Fatalf("counts offered = %v, want 1 and 2 (X = 0 is a no-op; X = 3 leaves no Mountain for {R})", counts)
	}
	for _, p := range got {
		if p.XValue != len(p.SacrificeIDs) {
			t.Errorf("move = %+v: X must be the sacrifice count", p)
		}
	}
	dispatchAll(t, g, seat.ID, moves)
}
