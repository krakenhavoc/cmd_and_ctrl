package legal_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// pawprint_enumerator_test.go — #4786, CR 700.2i: the move list for a
// pawprint spell in hand offers only selections within its {P} budget,
// and every one it offers is a cast the engine accepts (#544).

const oracleSeasonOfLoss = "31b5d0ad-3a4d-4f43-b0a9-c9f7a362680f"

func TestSeasonOfLossIsOfferedOnlyWithinItsPawprints(t *testing.T) {
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	spell := handCard(active, game.Card{
		Name: "Season of Loss", TypeLine: "Sorcery",
		OracleID: oracleSeasonOfLoss, ManaCost: "{3}{B}{B}",
	})
	lands(g, active, "Swamp", "Swamp", 5)
	advanceTo(t, g, game.StepPrecombatMain)

	moves := legal.EnumerateFor(g, active.ID)
	combos := explosiveDerailmentModes(castMovesFor(moves, spell), spell)
	pawprints := map[rune]int{'0': 1, '1': 2, '2': 3}
	for key := range combos {
		n := 0
		for _, r := range strings.ReplaceAll(key, ",", "") {
			n += pawprints[r]
		}
		if n > 5 {
			t.Errorf("offered modes %q, which spend %d {P}", key, n)
		}
	}
	// The per-source budget cuts the list (ADR 0122 §6.2), so which
	// mixes survive it is the walk's order; the {P}{P}{P} bullet alone
	// and five of the {P} one always fit.
	if !combos["2"] || !combos["0,0,0,0,0"] {
		t.Errorf("the {P}{P}{P} bullet and five {P} bullets fit and should be offered; combos: %v", combos)
	}
	dispatchAll(t, g, active.ID, castMovesFor(moves, spell))
}
