package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sacrifice_half_view_test.go — #5075: a "sacrifice half the lands you
// control, rounded up" additional cost (Tectonic Split) ships
// sacrifice_options listing the viewer's lands, with min = max = half of
// them rounded up, so the picker demands what CastSpell will.

const tectonicSplitViewOracle = "b71c3c1e-f927-4192-ac3a-cf6a442ffd00"

func TestSacrificeHalfCostShipsHalfTheLandsRoundedUp(t *testing.T) {
	g := buildActiveGame(t)
	me := g.Seats[0]
	for i := 0; i < 5; i++ {
		pushSacNCard(g, me.ID, game.Card{Name: "Forest", TypeLine: "Basic Land — Forest"})
	}
	pushSacNCard(g, g.Seats[1].ID, game.Card{Name: "Their Forest", TypeLine: "Basic Land — Forest"})
	pushSacNCard(g, me.ID, game.Card{Name: "Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2})

	split := game.NewCard("Tectonic Split", me.ID)
	split.OracleID = tectonicSplitViewOracle
	split.TypeLine = "Enchantment"
	split.ManaCost = "{4}{G}{G}"
	me.Hand.PushTop(split)
	v := ViewOfGame(g)
	var hand *CardView
	for i := range v.Seats[0].Hand.Cards {
		if v.Seats[0].Hand.Cards[i].InstanceID == split.InstanceID.String() {
			hand = &v.Seats[0].Hand.Cards[i]
		}
	}
	if hand == nil || hand.AdditionalCost == nil || hand.AdditionalCost.SacrificeOptions == nil {
		t.Fatalf("additional_cost.sacrifice_options absent: %+v", hand)
	}
	opts := hand.AdditionalCost.SacrificeOptions
	if opts.Min != 3 || opts.Max != 3 {
		t.Errorf("min/max = %d/%d, want 3/3", opts.Min, opts.Max)
	}
	if len(opts.Cards) != 5 || opts.All {
		t.Errorf("cards = %v (all %v), want the five lands to choose from", opts.Cards, opts.All)
	}
}
