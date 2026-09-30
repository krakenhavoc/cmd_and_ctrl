package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// either_cost_view_test.go — ADR 0100 §4: a card with an either/or
// additional cost ships `additional_cost.branches`, each in the shape of
// the cost itself plus its `key` and whether the viewer could pay it
// right now (`payable`, the predicate CastSpell asks). A card whose
// every branch is unpayable is not castable here (CR 601.2h).

const eitherViewOracle = "test-either-view"

func installEitherForView(t *testing.T) {
	t.Helper()
	prev := game.CatalogAdditionalCost
	cost := &game.AdditionalCost{
		Label: "Sacrifice an artifact or discard a card",
		Either: []game.AdditionalCost{
			{Key: "sacrifice", Label: "Sacrifice an artifact", Sacrifice: &game.TargetSpec{
				Label: "an artifact", Zones: []game.ZoneKind{game.ZoneBattlefield}, Min: 1, Max: 1,
				CardOK: func(_ *game.Game, _ uuid.UUID, c game.Card, _ game.ZoneKind) bool { return c.HasCardType("artifact") },
			}},
			{Key: "discard", Label: "Discard a card", DiscardCards: 1},
			{Key: "mana", Label: "Pay {5}", ManaCost: "{5}"},
			{Key: "life", Label: "Pay 3 life", PayLife: 3},
		},
	}
	game.CatalogAdditionalCost = func(key string) *game.AdditionalCost {
		if key == eitherViewOracle {
			return cost
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogAdditionalCost = prev })
}

func TestEitherCostBranchesAreProjected(t *testing.T) {
	installEitherForView(t)
	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[0]
	me.Hand.Cards = nil
	spell := game.NewCard("Test Answers", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{1}{R}"
	spell.OracleID = eitherViewOracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	// Nothing to sacrifice, nothing else in hand, 40 life: the mana and
	// life branches are payable, the other two are not.
	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	c := handCardView(t, v, 0, spell.InstanceID)
	ac := c.AdditionalCost
	if ac == nil || len(ac.Branches) != 4 {
		t.Fatalf("additional_cost = %+v, want four branches", ac)
	}
	want := []struct {
		key     string
		payable bool
	}{{"sacrifice", false}, {"discard", false}, {"mana", true}, {"life", true}}
	for i, w := range want {
		b := ac.Branches[i]
		if b.Key != w.key || b.Payable != w.payable {
			t.Errorf("branch %d = key %q payable %v, want %q %v", i, b.Key, b.Payable, w.key, w.payable)
		}
	}
	if ac.Branches[0].SacrificeOptions == nil || len(ac.Branches[0].SacrificeOptions.Cards) != 0 {
		t.Errorf("sacrifice branch options = %+v, want present and empty", ac.Branches[0].SacrificeOptions)
	}
	if ac.Branches[1].DiscardCards != 1 || ac.Branches[2].ManaCost != "{5}" || ac.Branches[3].PayLife != 3 {
		t.Errorf("branch components = %+v", ac.Branches)
	}
	if ac.DiscardCards != 0 || ac.SacrificeOptions != nil || ac.Label == "" {
		t.Errorf("the branched cost itself = %+v, want only its label", ac)
	}
	if !castableNow(g, me.ID, spell, game.ZoneHand, nil, "", []*game.AlternativeCost{nil}) {
		t.Error("castable_here = false with two payable branches")
	}

	// At 2 life the life branch closes (CR 119.4); the mana branch keeps
	// the card castable, because mana is never asked (CR 601.2g).
	me.Life = 2
	v = FilterViewFor(ViewOfGame(g), me.ID.String())
	c = handCardView(t, v, 0, spell.InstanceID)
	if c.AdditionalCost.Branches[3].Payable {
		t.Error("the life branch is payable at 2 life")
	}
	if !castableNow(g, me.ID, spell, game.ZoneHand, nil, "", []*game.AlternativeCost{nil}) {
		t.Error("castable_here = false with the mana branch open")
	}
}

func TestEitherCostWithNoPayableBranchIsNotCastable(t *testing.T) {
	prev := game.CatalogAdditionalCost
	cost := &game.AdditionalCost{
		Label: "Discard a card or pay 3 life",
		Either: []game.AdditionalCost{
			{Key: "discard", Label: "Discard a card", DiscardCards: 1},
			{Key: "life", Label: "Pay 3 life", PayLife: 3},
		},
	}
	game.CatalogAdditionalCost = func(key string) *game.AdditionalCost {
		if key == eitherViewOracle {
			return cost
		}
		if prev != nil {
			return prev(key)
		}
		return nil
	}
	t.Cleanup(func() { game.CatalogAdditionalCost = prev })

	g := buildActiveGame(t)
	for _, p := range g.Seats {
		if err := g.KeepHand(p.ID); err != nil {
			t.Fatalf("KeepHand: %v", err)
		}
	}
	me := g.Seats[0]
	me.Hand.Cards = nil
	me.Life = 2
	spell := game.NewCard("Test Triumph", me.ID)
	spell.TypeLine = "Instant"
	spell.ManaCost = "{1}{B}"
	spell.OracleID = eitherViewOracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)
	// castableNow is the stamp behind castable_here on a graveyard,
	// library-top or exile card; a hand card's castability is the legal
	// move list, which offers no branch it cannot pay.
	if castableNow(g, me.ID, spell, game.ZoneHand, nil, "", []*game.AlternativeCost{nil}) {
		t.Fatal("castable_here = true with no payable branch")
	}
	// A card to discard opens the discard branch.
	filler := game.NewCard("Filler", me.ID)
	filler.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(filler)
	if !castableNow(g, me.ID, spell, game.ZoneHand, nil, "", []*game.AlternativeCost{nil}) {
		t.Fatal("castable_here = false with the discard branch open")
	}
}
