package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tutors_test.go — the "shuffle, then put that card on top" cycle.
//
// The assertions that earn their keep are about WHERE and WHEN, not
// about whether a search happened:
//
//   - the card does not go to hand. These cost one mana because
//     collecting the card costs a draw step; a to-hand search is a
//     strictly better card than the one printed.
//   - the placement happens AFTER the shuffle. A shuffle that ran
//     last would put the tutored card back into the deck at random,
//     which is the bug this ordering exists to avoid — and it is
//     invisible unless the test seeds a library big enough for a
//     shuffle to actually move things.
//   - "reveal it" reaches the whole table. Enlightened Tutor tells
//     every opponent what your next draw is, and that is a real cost
//     of the card. Imperial Seal pays 2 life to skip it.

const (
	enlightenedTutorOracle = "c5229c17-b7be-4b05-b683-f2277edc4849"
	worldlyTutorOracle     = "e8863518-0bfa-49c3-8c6e-6c9116a81051"
	mysticalTutorOracle    = "fb81f95c-70f8-4eb7-8d15-15d0ae23ec03"
	imperialSealOracle     = "16cd0b90-f70c-4efa-b252-8de8784ef9a3"
	beseechOracle          = "cf94cafc-527e-4b27-8a28-7807435aaccf"
)

// tutorTopOfLibrary casts a put-on-top tutor, answers the search with
// `needle`, and returns the card now sitting on top.
//
// The answer is conditional because the engine skips the prompt when
// the predicate leaves no decision to make — a library with exactly
// one matching card auto-takes it. Both paths have to reach the same
// place, which is itself worth exercising: the put-on-top runs in
// finishSearchLocked, so it must fire whether the pick came from a
// prompt or from the no-decision shortcut.
func tutorTopOfLibrary(t *testing.T, g *game.Game, p *game.Player,
	name, typeLine, oracle string, needle uuid.UUID,
) game.Card {
	t.Helper()
	castCatalogSpell(t, g, name, typeLine, oracle, nil)
	passPriorityAroundTable(t, g)
	if searchChoiceFor(g, p.ID) != nil {
		answerSearchByID(t, g, p.ID, needle)
	}
	top, err := p.Library.Top()
	if err != nil {
		t.Fatalf("%s: library is empty after the tutor", name)
	}
	return top
}

func TestEnlightenedTutorPutsAnArtifactOnTopAndRevealsIt(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := pushLibraryCardForTest(me, game.Card{
		Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}",
	})
	// A creature that must NOT be offered — the predicate is the card.
	pushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	libBefore := me.Library.Size()
	handBefore := me.Hand.Size()

	top := tutorTopOfLibrary(t, g, me, "Enlightened Tutor", "Instant", enlightenedTutorOracle, needle)

	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored Sol Ring", top.Name)
	}
	if me.Hand.Contains(needle) {
		t.Error("the tutored card went to hand; this tutor puts it on top")
	}
	if me.Hand.Size() != handBefore {
		t.Errorf("hand is %d, want %d", me.Hand.Size(), handBefore)
	}
	if me.Library.Size() != libBefore {
		t.Errorf("library is %d, want %d — the card never leaves it", me.Library.Size(), libBefore)
	}
	// "Reveal it" is a real cost: the table learns your next draw.
	if !top.IsKnownTo(opp.ID) {
		t.Error("an opponent does not know the revealed card; Enlightened Tutor says reveal it")
	}
	if !top.IsKnownTo(me.ID) {
		t.Error("the searcher does not know their own tutored card")
	}
}

func TestWorldlyTutorFindsACreature(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := pushLibraryCardForTest(me, game.Card{
		Name: "Craterhoof Behemoth", TypeLine: "Creature — Beast",
	})

	top := tutorTopOfLibrary(t, g, me, "Worldly Tutor", "Instant", worldlyTutorOracle, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored creature", top.Name)
	}
}

func TestMysticalTutorFindsAnInstantOrSorcery(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := pushLibraryCardForTest(me, game.Card{
		Name: "Counterspell", TypeLine: "Instant",
	})

	top := tutorTopOfLibrary(t, g, me, "Mystical Tutor", "Instant", mysticalTutorOracle, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored instant", top.Name)
	}
}

// TestImperialSealIsPrivateAndCostsLife — the one card in the cycle
// with no "reveal it". Nobody but the controller learns the top card,
// and the 2 life is what buys that.
func TestImperialSealIsPrivateAndCostsLife(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := pushLibraryCardForTest(me, game.Card{Name: "Anything At All", TypeLine: "Sorcery"})
	lifeBefore := me.Life

	top := tutorTopOfLibrary(t, g, me, "Imperial Seal", "Sorcery", imperialSealOracle, needle)

	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored card", top.Name)
	}
	if me.Life != lifeBefore-2 {
		t.Errorf("life is %d, want %d", me.Life, lifeBefore-2)
	}
	if top.IsKnownTo(opp.ID) {
		t.Error("an opponent knows the top card; Imperial Seal does not say reveal it")
	}
	if !top.IsKnownTo(me.ID) {
		t.Error("the searcher does not know their own tutored card")
	}
}

// TestPutOnTopSurvivesTheShuffle — the ordering assertion. The
// library is big enough that a shuffle would almost certainly move
// the needle off the top; if the placement ran before the shuffle
// instead of after, this fails.
func TestPutOnTopSurvivesTheShuffle(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 30; i++ {
		pushLibraryCardForTest(me, game.Card{Name: "Chaff", TypeLine: "Sorcery"})
	}
	needle := pushLibraryCardForTest(me, game.Card{Name: "The Needle", TypeLine: "Instant"})

	top := tutorTopOfLibrary(t, g, me, "Mystical Tutor", "Instant", mysticalTutorOracle, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q — the placement ran before the shuffle", top.Name)
	}
}

// --- to-hand tutors -------------------------------------------------
//
// Fabricate is covered by batch03_test.go's
// TestFabricateTutorsAnArtifactToHand — it landed on main from the
// card roadmap while this branch was open, and its version of the
// card is the incumbent.

// TestBeseechTheQueenIsCappedByLandCount — the card's whole design is
// the ceiling, so the test that matters is the one where the ceiling
// bites.
func TestBeseechTheQueenIsCappedByLandCount(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	// Two lands, so mana value 2 is findable and 5 is not.
	pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	pushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	cheap := pushLibraryCardForTest(me, game.Card{
		Name: "Two Drop", TypeLine: "Creature — Human", ManaCost: "{1}{U}",
	})
	expensive := pushLibraryCardForTest(me, game.Card{
		Name: "Five Drop", TypeLine: "Creature — Giant", ManaCost: "{4}{U}",
	})

	castCatalogSpell(t, g, "Beseech the Queen", "Sorcery", beseechOracle, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	offered := map[uuid.UUID]bool{}
	for _, id := range c.SearchCards {
		offered[id] = true
	}
	if !offered[cheap] {
		t.Error("a mana-value-2 card was not offered with two lands out")
	}
	if offered[expensive] {
		t.Error("a mana-value-5 card was offered with only two lands out")
	}

	answerSearchByID(t, g, me.ID, cheap)
	if !me.Hand.Contains(cheap) {
		t.Error("the tutored card is not in hand")
	}
}
