package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// discard_hand_view_test.go — #1600. A "Discard your hand" cost names
// no card, so the wire carries nothing for it: no discard_cost_n, no
// options, nothing for the client to open a picker on. "Activate only
// as an instant" reaches the client the way every other activation
// condition does, as condition_unmet, off the same closure the engine
// gates on.

func TestDiscardYourHandManaAbilityViewStampsNoPickerAndTheInstantWindow(t *testing.T) {
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[g.Turn.ActiveSeat]
	if me.Hand.Size() == 0 {
		t.Fatal("test premise: a full hand")
	}
	id := uuid.New()
	g.Battlefield.PushTop(game.Card{
		InstanceID: id,
		Name:       "Test Diamond",
		TypeLine:   "Artifact",
		OracleID:   "00000000-0000-0000-0000-000000001600",
		Owner:      me.ID,
		Controller: me.ID,
		ManaAbilities: []game.ManaAbilityShape{{
			SacrificeCost: true,
			DiscardCards:  &game.DiscardCost{Hand: true, Label: "your hand"},
			Produced:      "{R}{R}{R}",
			Label:         "Discard your hand, Sacrifice this artifact: Add {R}{R}{R}. Activate only as an instant.",
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return g.InstantWindowOpenForEffect(controller)
			},
		}},
	})
	g.BumpLayerVersionForTest()

	ma := controllerFrameCard(t, g, id).ManaAbilities[0]
	if ma.DiscardCostN != 0 || ma.DiscardCostLabel != "" || len(ma.DiscardCostOptions) != 0 {
		t.Errorf("a hand clause was published as a pick: n=%d label=%q options=%v",
			ma.DiscardCostN, ma.DiscardCostLabel, ma.DiscardCostOptions)
	}
	if !ma.SacrificeCost {
		t.Error("the sacrifice-this half is missing from the view")
	}
	if ma.ConditionUnmet {
		t.Error("condition_unmet with priority and nothing open")
	}

	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if ma := controllerFrameCard(t, g, id).ManaAbilities[0]; !ma.ConditionUnmet {
		t.Error("condition_unmet is clear while another seat holds priority")
	}
}
