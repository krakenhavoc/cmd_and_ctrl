package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// face_down_devotion_test.go — #4761. CR 700.5: devotion counts the
// mana symbols among the mana costs of permanents you control. CR
// 708.2a: a face-down permanent has no mana cost, so it adds nothing,
// whatever card it hides. Card.ManaCost still holds the hidden card's
// cost, and before #4761 both devotion reads parsed it: Gray Merchant
// drained for symbols nobody could see, a God turned on early, and what
// either did told the table the colours of the face-down card.

// faceDownHiding puts a face-down permanent of `kind` onto the
// battlefield under owner, hiding a card that costs `cost`.
func faceDownHiding(g *game.Game, owner *game.Player, kind game.FaceDownKind, cost string) uuid.UUID {
	c := game.Card{
		Name: "Hidden", TypeLine: "Creature — Horror", ManaCost: cost, Power: 5, Toughness: 5,
		KnownBy: map[uuid.UUID]bool{owner.ID: true},
	}
	c.SetFaceDown(kind)
	return apaPush(g, owner.ID, owner.ID, c)
}

// TestDevotionCountsNoSymbolsForAFaceDownPermanent is both readers, for
// every way the engine makes a face-down permanent. The face-up {B}{R}
// beside it is the control: it counts, so the zero is the face-down
// permanent's and not an empty board's.
func TestDevotionCountsNoSymbolsForAFaceDownPermanent(t *testing.T) {
	for _, kind := range []game.FaceDownKind{
		game.FaceDownManifested, game.FaceDownMorphed, game.FaceDownDisguised,
		game.FaceDownCloaked, game.FaceDownTurned,
	} {
		t.Run(string(kind), func(t *testing.T) {
			g := newCatalogGame(t)
			me := g.Seats[0]
			faceDownHiding(g, me, kind, "{B}{B}{R}{R}{B/R}")
			apaPush(g, me.ID, me.ID, game.Card{Name: "Face Up", TypeLine: "Enchantment", ManaCost: "{B}{R}"})

			if got := devotionTo(g, me.ID, "B"); got != 1 {
				t.Errorf("devotion to black = %d, want 1: the face-down permanent has no mana cost (CR 708.2a)", got)
			}
			if got := devotionTo(g, me.ID, "R"); got != 1 {
				t.Errorf("devotion to red = %d, want 1", got)
			}
			if got := devotionToColors(g, me.ID, "B", "R"); got != 2 {
				t.Errorf("devotion to black and red = %d, want 2", got)
			}
		})
	}
}

// TestGrayMerchantDoesNotDrainForAFaceDownPermanent: Gary's own cost is
// left off the seeded card, so the drain is exactly the board's devotion
// to black, and the only black symbols on the board are hidden.
func TestGrayMerchantDoesNotDrainForAFaceDownPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	faceDownHiding(g, me, game.FaceDownManifested, "{4}{B}{B}")
	before := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		before[p.ID] = p.Life
	}

	castCatalogSpell(t, g, "Gray Merchant of Asphodel", "Creature — Zombie", b02bGrayMerchantOracle, nil)
	passPriorityAroundTable(t, g)

	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Errorf("seat %v life %d → %d: Gray Merchant drained for the black symbols of a face-down card",
				p.ID, before[p.ID], p.Life)
		}
	}
}

// TestPurphorosIsNotTurnedOnByAFaceDownPermanent is the one-colour God
// (godUnlessDevotion, devotionTo): one red symbol of its own, and four
// hidden ones that do not count.
func TestPurphorosIsNotTurnedOnByAFaceDownPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	p := b27Push(g, me.ID, "Purphoros, God of the Forge", "Legendary Enchantment Creature — God",
		gl1bPurphorosOracle, "{3}{R}", 6, 5, "R")
	faceDownHiding(g, me, game.FaceDownManifested, "{R}{R}{R}{R}")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if b39IsCreature(t, g, p) {
		t.Error("Purphoros is a creature at devotion 1: it counted the four red symbols of a face-down card")
	}
}

// TestMogisIsNotTurnedOnByAFaceDownPermanent is the two-colour God
// (godUnlessDevotionToColors, devotionToColors): 2 + 3 face up is five,
// and the hidden four would make nine.
func TestMogisIsNotTurnedOnByAFaceDownPermanent(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	mogis := b27Push(g, me.ID, "Mogis, God of Slaughter", "Legendary Enchantment Creature — God",
		s58Mogis, "{2}{B}{R}", 7, 5, "B", "R")
	b27Push(g, me.ID, "Black Thing", "Enchantment", "", "{B}{B}{B}", 0, 0, "B")
	faceDownHiding(g, me, game.FaceDownMorphed, "{B}{B}{R}{R}")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if b39IsCreature(t, g, mogis) {
		t.Error("Mogis is a creature at devotion 5: it counted the symbols of a face-down card")
	}

	// Two more face-up symbols make seven, so the gate itself still works.
	b27Push(g, me.ID, "Red Thing", "Enchantment", "", "{R}{R}", 0, 0, "R")
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if !b39IsCreature(t, g, mogis) {
		t.Error("devotion 7 face up: Mogis is a creature")
	}
}
