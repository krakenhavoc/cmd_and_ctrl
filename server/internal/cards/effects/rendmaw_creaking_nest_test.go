package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const rendmawOracle = "2207fb3f-45a1-401b-99e9-f75723495c3f"

func TestRendmawCreakingNestIsRegisteredFull(t *testing.T) {
	spec, ok := Lookup(rendmawOracle)
	if !ok || spec.Name != "Rendmaw, Creaking Nest" || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Fatalf("registration: %+v", spec)
	}
}

func rendmawBirds(g *game.Game) []game.Card {
	var out []game.Card
	for _, c := range g.Battlefield.Cards {
		if c.Name == "Bird" && c.IsToken() {
			out = append(out, c)
		}
	}
	return out
}

func TestRendmawMakesEachPlayerAGoadedTappedBirdOnEntering(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	rfdCast(t, g, "Rendmaw, Creaking Nest", "Legendary Artifact Creature — Scarecrow", rendmawOracle, 5, 5)
	passPriorityAroundTable(t, g)

	birds := rendmawBirds(g)
	if len(birds) != len(g.Seats) {
		t.Fatalf("%d Birds, want one per player (%d)", len(birds), len(g.Seats))
	}
	owners := map[string]bool{}
	for _, b := range birds {
		owners[b.Controller.String()] = true
		if !b.Tapped || b.Power != 2 || b.Toughness != 2 || !b.HasColor("B") {
			t.Errorf("Bird = %d/%d tapped=%v %v, want a tapped 2/2 black one", b.Power, b.Toughness, b.Tapped, b.Colors)
		}
		if !b.IsGoadedBy(me.ID) {
			t.Errorf("a Bird of %s is not goaded by Rendmaw's controller", b.Controller)
		}
	}
	if len(owners) != len(g.Seats) {
		t.Errorf("Birds belong to %d players, want %d", len(owners), len(g.Seats))
	}
}

func TestRendmawTriggersOnPlayingACardWithTwoCardTypesOnly(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	b12Push(g, me.ID, "Rendmaw, Creaking Nest", "Legendary Artifact Creature — Scarecrow", rendmawOracle, 5, 5)

	// A creature spell has one card type.
	tdCastFromHand(t, g, game.Card{Name: "Plain Bear", TypeLine: "Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := len(rendmawBirds(g)); n != 0 {
		t.Fatalf("a one-type spell made %d Birds", n)
	}
	// An artifact creature has two.
	tdCastFromHand(t, g, game.Card{Name: "Clockwork Bear", TypeLine: "Artifact Creature — Bear", Power: 2, Toughness: 2, ManaCost: "{1}"}, game.CastSpellParams{})
	passPriorityAroundTable(t, g)
	if n := len(rendmawBirds(g)); n != len(g.Seats) {
		t.Fatalf("a two-type spell made %d Birds, want %d", n, len(g.Seats))
	}
}
