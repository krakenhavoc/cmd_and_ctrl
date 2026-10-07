package protocol

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reveal_cost_view_test.go — ADR 0100 amendment 2026-10-07: a reveal or
// behold branch ships `reveal` / `behold` and `reveal_options`, the
// engine's own candidate walk, never including the spell itself.

const revealViewOracle = "test-reveal-view"

func TestRevealBranchProjectsItsOptions(t *testing.T) {
	prev := game.CatalogAdditionalCost
	cost := &game.AdditionalCost{
		Label: "Behold a Merfolk or pay {2}",
		Either: []game.AdditionalCost{
			{Key: "behold", Label: "Behold a Merfolk", Reveal: &game.RevealCost{Subtype: "Merfolk", Behold: true}},
			{Key: "mana", Label: "Pay {2}", ManaCost: "{2}"},
		},
	}
	game.CatalogAdditionalCost = func(key string) *game.AdditionalCost {
		if key == revealViewOracle {
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
	spell := game.NewCard("Test Mentor", me.ID)
	spell.TypeLine = "Creature — Merfolk Wizard"
	spell.ManaCost = "{1}{U}"
	spell.OracleID = revealViewOracle
	spell.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(spell)

	// The only Merfolk is the spell itself: the behold branch is closed.
	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	ac := handCardView(t, v, 0, spell.InstanceID).AdditionalCost
	if ac == nil || len(ac.Branches) != 2 {
		t.Fatalf("additional_cost = %+v, want two branches", ac)
	}
	b := ac.Branches[0]
	if !b.Reveal || !b.Behold || b.RevealOptions == nil || len(b.RevealOptions.Cards) != 0 || b.Payable {
		t.Errorf("behold branch with only the spell = %+v, want reveal+behold, present-and-empty options, not payable", b)
	}
	if ac.Branches[1].Reveal || !ac.Branches[1].Payable {
		t.Errorf("mana branch = %+v, want a payable non-reveal branch", ac.Branches[1])
	}

	// A hand Merfolk and a Merfolk on the table both become options.
	pal := game.NewCard("Merfolk Pal", me.ID)
	pal.TypeLine = "Creature — Merfolk"
	pal.KnownBy = map[uuid.UUID]bool{me.ID: true}
	me.Hand.PushTop(pal)
	scout := game.NewCard("Merfolk Scout", me.ID)
	scout.TypeLine = "Creature — Merfolk"
	scout.Controller = me.ID
	g.Battlefield.PushTop(scout)
	v = FilterViewFor(ViewOfGame(g), me.ID.String())
	b = handCardView(t, v, 0, spell.InstanceID).AdditionalCost.Branches[0]
	if !b.Payable || b.RevealOptions == nil || len(b.RevealOptions.Cards) != 2 {
		t.Fatalf("behold branch = %+v, want two options (hand card then permanent)", b)
	}
	if b.RevealOptions.Cards[0] != pal.InstanceID.String() || b.RevealOptions.Cards[1] != scout.InstanceID.String() {
		t.Errorf("options = %v, want hand first then the permanent", b.RevealOptions.Cards)
	}
}
