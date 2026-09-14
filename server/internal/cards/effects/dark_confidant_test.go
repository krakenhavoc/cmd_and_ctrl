package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dark_confidant_test.go — the upkeep flip, and the three things
// about it that are easy to get wrong:
//
//   - it REVEALS, so the whole table learns the card;
//   - it does not DRAW, so nothing watching draws fires;
//   - the life is the flipped card's mana value, which is zero for a
//     land and is charged after the card has reached hand.

const darkConfidantOracle = "2068185c-1b50-47d0-aa3f-bf505d199428"

// seedLibraryTop puts one card with a real mana cost on top of the
// player's library. seedLibrary makes costless Sorceries, which is
// the wrong shape for a card whose whole clause is "equal to its mana
// value".
func seedLibraryTop(p *game.Player, name, typeLine, manaCost string) uuid.UUID {
	id := uuid.New()
	p.Library.PushTop(game.Card{
		InstanceID: id,
		Name:       name,
		TypeLine:   typeLine,
		ManaCost:   manaCost,
		Owner:      p.ID,
		Controller: p.ID,
	})
	return id
}

// runDarkConfidantUpkeep puts a Confidant on seat 1's battlefield,
// walks the table to that seat's upkeep and resolves the trigger.
func runDarkConfidantUpkeep(t *testing.T, g *game.Game) *game.Player {
	t.Helper()
	owner := g.Seats[1]
	pushPermanentForTest(g, owner.ID, "Dark Confidant", darkConfidantOracle, "Creature — Human Wizard")
	advanceToStepOf(t, g, 1, game.StepUpkeep)
	passPriorityAroundTable(t, g)
	return owner
}

// TestDarkConfidantFlipsTheTopCardAndCharges is the card.
func TestDarkConfidantFlipsTheTopCardAndCharges(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	// {2}{B}{B} — mana value 4, and distinct from the number of cards
	// moving so a wrong reading cannot coincidentally pass.
	flipped := seedLibraryTop(owner, "Gray Merchant of Asphodel", "Creature — Zombie", "{2}{B}{B}")

	handBefore := owner.Hand.Size()
	libBefore := owner.Library.Size()
	lifeBefore := owner.Life

	runDarkConfidantUpkeep(t, g)

	if !owner.Hand.Contains(flipped) {
		t.Fatal("the revealed card never reached hand")
	}
	if got := owner.Hand.Size(); got != handBefore+1 {
		t.Errorf("hand = %d, want %d", got, handBefore+1)
	}
	if got := owner.Library.Size(); got != libBefore-1 {
		t.Errorf("library = %d, want %d", got, libBefore-1)
	}
	if got := lifeBefore - owner.Life; got != 4 {
		t.Errorf("life lost = %d, want 4 (the flipped card's mana value)", got)
	}
}

// TestDarkConfidantRevealsToTheWholeTable is the #549 half: the flip
// is an announcement, not a private look. An opponent who cannot see
// a library must still be told what Bob just charged for.
func TestDarkConfidantRevealsToTheWholeTable(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	seedLibraryTop(owner, "Gray Merchant of Asphodel", "Creature — Zombie", "{2}{B}{B}")

	runDarkConfidantUpkeep(t, g)

	for _, seat := range g.Seats {
		names := revealedNames(revealsFor(g, seat.ID))
		if !containsName(names, "Gray Merchant of Asphodel") {
			t.Errorf("seat %s was not told what the Confidant flipped; window = %v", seat.Name, names)
		}
	}
}

// TestDarkConfidantDoesNotDraw is the clause a draw deck cares about
// most. "Reveal and put into your hand" is not a draw, so a draw
// payoff sitting beside the Confidant stays quiet — the Crawler here
// is the real catalog card, drained by nothing.
func TestDarkConfidantDoesNotDraw(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	seedLibraryTop(owner, "Gray Merchant of Asphodel", "Creature — Zombie", "{2}{B}{B}")
	pushPermanentForTest(g, owner.ID, "Psychosis Crawler", psychosisCrawlerOracle,
		"Artifact Creature — Phyrexian Horror")

	opponent := g.Seats[2]
	oppLifeBefore := opponent.Life

	runDarkConfidantUpkeep(t, g)

	if opponent.Life != oppLifeBefore {
		t.Errorf("an opponent lost %d life to Psychosis Crawler on a flip that is not a draw",
			oppLifeBefore-opponent.Life)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventDrawCard && ev.Actor == owner.ID {
			t.Fatal("the Confidant's flip emitted EventDrawCard; revealing is not drawing")
		}
	}
}

// TestDarkConfidantChargesNothingForALand is the mana-value-zero leg,
// and the reason a Bob deck runs a high land count.
func TestDarkConfidantChargesNothingForALand(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	land := seedLibraryTop(owner, "Island", "Basic Land — Island", "")
	lifeBefore := owner.Life

	runDarkConfidantUpkeep(t, g)

	if !owner.Hand.Contains(land) {
		t.Error("the land never reached hand")
	}
	if owner.Life != lifeBefore {
		t.Errorf("life %d -> %d on a land flip, want no change", lifeBefore, owner.Life)
	}
}

// TestDarkConfidantOnAnEmptyLibraryIsNotALoss — revealing is not
// drawing, so bottoming out this way does not arm CR 704.5b.
func TestDarkConfidantOnAnEmptyLibraryIsNotALoss(t *testing.T) {
	g := newCatalogGame(t)
	owner := g.Seats[1]
	owner.Library.Cards = nil
	handBefore := owner.Hand.Size()
	lifeBefore := owner.Life

	runDarkConfidantUpkeep(t, g)

	if owner.Hand.Size() != handBefore {
		t.Errorf("hand = %d, want %d — an empty library reveals nothing", owner.Hand.Size(), handBefore)
	}
	if owner.Life != lifeBefore {
		t.Errorf("life %d -> %d, want no change", lifeBefore, owner.Life)
	}
	if owner.Eliminated {
		t.Error("the Confidant's controller lost to an empty library; revealing is not drawing")
	}
}
