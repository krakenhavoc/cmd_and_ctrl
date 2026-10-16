package heuristic_test

import (
	"context"
	"math"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// library_exile_cost_test.go — #3966: legal.MoveCost.LibraryExile
// reaching the policy. An "Exile the top N cards of your library" cost
// names no card in the move's params, so before the count rode the move
// a bot paid its library for Seasoned Tactician without weighing it. It
// is now priced as a mill without the graveyard fuel, and declined under
// the same DredgeLibraryFloor.

const libraryExileLabel = "Tactician Stand-in: {1}, Exile the top four cards of your library: Prevent damage."

func libraryExileSource(id string) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Tactician Stand-in", TypeLine: "Creature",
		Owner: seatID(0).String(), Controller: seatID(0).String(), KnownByYou: true,
		ActivatedAbilities: []protocol.ActivatedAbilityView{{Index: 0, ManaCost: "{1}", LibraryExileCostN: 4}},
	}
}

// libraryExileRank ranks pass, the activation as if it were free, the
// same activation exiling four off the library and (for comparison) one
// milling four, for seat 0 with seat opts applied.
func libraryExileRank(t *testing.T, p *heuristic.Policy, opts ...seatOpt) (free, exiled, milled heuristic.Candidate) {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, opts...), newSeat(1)},
		withBattlefield(append([]protocol.CardView{libraryExileSource(cardID(1))}, lands(4)...)...))
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), libraryExileLabel+" (free)", nil),
		activateMove(t, 0, cardID(1), libraryExileLabel, &legal.MoveCost{LibraryExile: 4}),
		activateMove(t, 0, cardID(1), libraryExileLabel+" (milled)", &legal.MoveCost{Mill: 4}),
	)
	for _, c := range p.Rank(context.Background(), in) {
		switch c.Index {
		case 1:
			free = c
		case 2:
			exiled = c
		case 3:
			milled = c
		}
	}
	return free, exiled, milled
}

// An exile that would leave the library under the floor is declined
// whatever the activation buys: below passing, with the reason said.
func TestLibraryExileCostUnderTheLibraryFloorIsDeclined(t *testing.T) {
	free, exiled, _ := libraryExileRank(t, heuristic.New(), withLibraryCount(12))
	if exiled.Value >= 0 {
		t.Fatalf("exiling 4 of a 12-card library is valued %.2f — want it below passing", exiled.Value)
	}
	if !strings.Contains(exiled.Reason, "too short") {
		t.Errorf("reason %q does not say the library is too short", exiled.Reason)
	}
	if free.Value <= exiled.Value {
		t.Errorf("the free activation (%.2f) should outrank the declined exile (%.2f)", free.Value, exiled.Value)
	}
}

// Above the floor each exiled card costs the library card it was and
// buys nothing back: the move is worth less than the free one, the
// graveyard does not change it, and a mill of the same cards into a
// castable graveyard is worth more.
func TestLibraryExileCostIsAMillWithoutTheFuel(t *testing.T) {
	deadLands := withGraveyard(land(cardID(20), 0), land(cardID(21), 0))
	fuel := func() protocol.CardView {
		c := creature(cardID(22), 0, "Flashback Stand-in", 3, 3)
		c.ManaCost = "{2}{B}"
		c.CastableHere = true
		return c
	}
	live := withGraveyard(fuel())

	free, overLands, _ := libraryExileRank(t, heuristic.New(), withLibraryCount(60), deadLands)
	if overLands.Value >= free.Value {
		t.Errorf("the exile is valued %.2f against %.2f free — want it priced", overLands.Value, free.Value)
	}
	freeLive, overFuel, milledLive := libraryExileRank(t, heuristic.New(), withLibraryCount(60), live)
	if d, dFree := freeLive.Value-overFuel.Value, free.Value-overLands.Value; math.Abs(d-dFree) > 1e-9 {
		t.Errorf("the exile's price moved with the graveyard: %.4f over fuel, %.4f over dead lands", d, dFree)
	}
	if milledLive.Value <= overFuel.Value {
		t.Errorf("milling into a castable graveyard (%.2f) should be worth more than exiling the same cards (%.2f)",
			milledLive.Value, overFuel.Value)
	}
}

// BaselineConfig reads the exile as free, as the heuristic did before,
// floor or not.
func TestBaselineConfigReadsALibraryExileCostAsFree(t *testing.T) {
	free, exiled, _ := libraryExileRank(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), withLibraryCount(12))
	if exiled.Value != free.Value {
		t.Errorf("baseline values the exile %.2f against %.2f free — want them equal", exiled.Value, free.Value)
	}
}
