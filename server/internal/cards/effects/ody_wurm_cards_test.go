package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ody_wurm_cards_test.go — the 6/6 green Wurm row and the two Odyssey
// block flashback cards that make it (#5789; Roar of the Wurm from
// Odyssey, #5613, and Crush of Wurms from Judgment, #5611). Each test
// plays the real catalog entry.

const (
	odyWurmRoarOracle  = "b8b9e6fd-b3ed-4fbc-8753-74018fa33caa"
	odyWurmCrushOracle = "6c12cd05-b673-453c-9f6b-f3aa022467c1"
)

var odyWurmCards = []struct {
	name, oracle, flashback string
	tokens                  int
}{
	{"Roar of the Wurm", odyWurmRoarOracle, "{3}{G}", 1},
	{"Crush of Wurms", odyWurmCrushOracle, "{9}{G}{G}{G}", 3},
}

func TestOdyWurmTokenRowIsAPlainSixSixGreenWurm(t *testing.T) {
	w := TokenCard("6/6 green Wurm")
	if w.Name != "Wurm" || w.TypeLine != "Token Creature — Wurm" || w.Power != 6 || w.Toughness != 6 {
		t.Errorf("row = %q %q %d/%d, want a Wurm token creature 6/6", w.Name, w.TypeLine, w.Power, w.Toughness)
	}
	if len(w.Colors) != 1 || w.Colors[0] != "G" || len(w.Keywords) != 0 {
		t.Errorf("row colours %v keywords %v, want green and no keywords", w.Colors, w.Keywords)
	}
}

func TestOdyWurmCardsAreRegisteredFull(t *testing.T) {
	for _, c := range odyWurmCards {
		spec, ok := Lookup(c.oracle)
		if !ok {
			t.Errorf("%s is not registered", c.name)
			continue
		}
		if spec.Name != c.name || spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
			t.Errorf("%s: name %q, completeness %s, caveats %v", c.name, spec.Name, spec.Completeness, spec.Caveats)
		}
	}
}

func TestOdyWurmCardsOfferTheirPrintedFlashbackFromTheGraveyard(t *testing.T) {
	for _, c := range odyWurmCards {
		t.Run(c.name, func(t *testing.T) {
			if !game.CardCastableFromZone(c.oracle, game.ZoneGraveyard) {
				t.Fatalf("%s does not open the graveyard", c.name)
			}
			offers := game.AlternativeCostsOfferedFromZone(c.oracle, game.ZoneGraveyard)
			if len(offers) != 1 || offers[0].Key != "flashback" {
				t.Fatalf("%s graveyard offers = %+v, want one flashback", c.name, offers)
			}
			if offers[0].ManaCost != c.flashback || !offers[0].ExileOnLeavingStack {
				t.Errorf("%s flashback pays %q (exile %v), want %q and exile", c.name, offers[0].ManaCost, offers[0].ExileOnLeavingStack, c.flashback)
			}
		})
	}
}

// The hand cast makes the printed number of Wurms, and the flashback
// cast, paid from a pool holding exactly the flashback price, makes as
// many again and exiles the card (odyB1Flashback checks the pool is
// spent and the card is in exile).
func TestOdyWurmCardsMakeSixSixGreenWurmsFromHandAndGraveyard(t *testing.T) {
	for _, c := range odyWurmCards {
		t.Run(c.name, func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[g.Turn.ActiveSeat]
			castCatalogSpell(t, g, c.name, "Sorcery", c.oracle, nil)
			passPriorityAroundTable(t, g)
			if got := b43TokensNamed(g, me.ID, "Wurm"); got != c.tokens {
				t.Fatalf("%d Wurm tokens after the hand cast, want %d", got, c.tokens)
			}

			id := seedGraveyardCard(t, g, c.name, "Sorcery", c.oracle)
			odyB1Flashback(t, g, id, c.flashback)
			if got := b43TokensNamed(g, me.ID, "Wurm"); got != 2*c.tokens {
				t.Fatalf("%d Wurm tokens after the flashback, want %d", got, 2*c.tokens)
			}
			if me.Graveyard.Contains(id) {
				t.Error("the flashed-back card went back to the graveyard")
			}
			for _, w := range g.Battlefield.Cards {
				if !w.IsToken() || w.Name != "Wurm" {
					continue
				}
				if p, tough := ptOf(t, g, w.InstanceID); p != 6 || tough != 6 || !w.HasColor("G") {
					t.Errorf("a Wurm token is %d/%d green=%v, want a green 6/6", p, tough, w.HasColor("G"))
				}
			}
		})
	}
}

// The graveyard path charges the flashback price, never the printed
// mana cost: Crush of Wurms' {6}{G}{G}{G} does not flash it back.
func TestOdyWurmCrushFlashbackIsNotPaidByTheHandCost(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	id := seedGraveyardCard(t, g, "Crush of Wurms", "Sorcery", odyWurmCrushOracle)
	dcMana(t, g, me.ID, odyB1Pool("{6}{G}{G}{G}"))
	if err := g.CastSpell(me.ID, id, game.CastSpellParams{
		FromZone: "graveyard", AlternativeCost: "flashback", Strict: true,
	}); err == nil {
		t.Fatal("Crush of Wurms flashed back for its hand cost")
	}
	if !me.Graveyard.Contains(id) {
		t.Error("a refused flashback moved Crush of Wurms out of the graveyard")
	}
	if got := b43TokensNamed(g, me.ID, "Wurm"); got != 0 {
		t.Errorf("%d Wurm tokens after a refused flashback, want 0", got)
	}
}
