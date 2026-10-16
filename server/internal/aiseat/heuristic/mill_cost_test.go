package heuristic_test

import (
	"context"
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// mill_cost_test.go — #3870: legal.MoveCost.Mill reaching the policy.
// A "Mill N cards" cost names no card in the move's params, so before
// the count rode the move a bot paid its library for Rot Farm Skeleton
// without weighing it. It is now priced at the dredge's rate for the
// same cards, against what the graveyard already holds, and declined
// when it would leave the library under DredgeLibraryFloor.

const millLabel = "Rot Farm Stand-in: {2}{B}{G}, Mill four cards: Return this card from your graveyard to the battlefield."

func millSource(id string) protocol.CardView {
	return protocol.CardView{
		InstanceID: id, Name: "Rot Farm Stand-in", TypeLine: "Artifact",
		Owner: seatID(0).String(), Controller: seatID(0).String(), KnownByYou: true,
		ActivatedAbilities: []protocol.ActivatedAbilityView{{Index: 0, ManaCost: "{2}{B}{G}", MillCost: 4}},
	}
}

// millRank ranks pass, the activation as if it were free, and the same
// activation milling four, for seat 0 with seat opts applied, and
// returns the two activations' candidates.
func millRank(t *testing.T, p *heuristic.Policy, opts ...seatOpt) (free, milled heuristic.Candidate) {
	t.Helper()
	v := newView([]protocol.PlayerView{newSeat(0, opts...), newSeat(1)},
		withBattlefield(append([]protocol.CardView{millSource(cardID(1))}, lands(4)...)...))
	in := input(0, v,
		passMove(0),
		activateMove(t, 0, cardID(1), millLabel+" (free)", nil),
		activateMove(t, 0, cardID(1), millLabel, &legal.MoveCost{Mill: 4}),
	)
	for _, c := range p.Rank(context.Background(), in) {
		switch c.Index {
		case 1:
			free = c
		case 2:
			milled = c
		}
	}
	return free, milled
}

// A mill that would leave the library under the floor is declined
// whatever the activation buys: below passing, with the reason said.
func TestMillCostUnderTheLibraryFloorIsDeclined(t *testing.T) {
	free, milled := millRank(t, heuristic.New(), withLibraryCount(12))
	if milled.Value >= 0 {
		t.Fatalf("milling 4 of a 12-card library is valued %.2f — want it below passing", milled.Value)
	}
	if !strings.Contains(milled.Reason, "library too short") {
		t.Errorf("reason %q does not say the library is too short", milled.Reason)
	}
	if free.Value <= milled.Value {
		t.Errorf("the free activation (%.2f) should outrank the declined mill (%.2f)", free.Value, milled.Value)
	}
}

// Above the floor the milled cards are priced as the dredge prices
// them: graveyard fuel less the library card each was. A graveyard the
// seat can cast from makes the same mill worth more than one of dead
// lands.
func TestMillCostIsPricedByWhatTheGraveyardWants(t *testing.T) {
	deadLands := withGraveyard(land(cardID(20), 0), land(cardID(21), 0))
	fuel := func() protocol.CardView {
		c := creature(cardID(22), 0, "Flashback Stand-in", 3, 3)
		c.ManaCost = "{2}{B}"
		c.CastableHere = true
		return c
	}
	live := withGraveyard(fuel())

	_, overLands := millRank(t, heuristic.New(), withLibraryCount(60), deadLands)
	_, overFuel := millRank(t, heuristic.New(), withLibraryCount(60), live)
	if overFuel.Value <= overLands.Value {
		t.Errorf("milling into a castable graveyard is valued %.2f, into dead lands %.2f — want it higher",
			overFuel.Value, overLands.Value)
	}
	free, milled := millRank(t, heuristic.New(), withLibraryCount(60), deadLands)
	if milled.Value == free.Value {
		t.Errorf("the mill is priced at nothing (%.2f both ways)", milled.Value)
	}
}

// BaselineConfig reads the mill as free, as the heuristic did before,
// floor or not.
func TestBaselineConfigReadsAMillCostAsFree(t *testing.T) {
	free, milled := millRank(t, heuristic.NewWithConfig(heuristic.BaselineConfig()), withLibraryCount(12))
	if milled.Value != free.Value {
		t.Errorf("baseline values the mill %.2f against %.2f free — want them equal", milled.Value, free.Value)
	}
}
