package teststo

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func TestEnlightenedTutorPutsAnArtifactOnTopAndRevealsIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Sol Ring", TypeLine: "Artifact", ManaCost: "{1}",
	})
	// A creature that must NOT be offered — the predicate is the card.
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	libBefore := me.Library.Size()
	handBefore := me.Hand.Size()

	top := effectstest.TutorTopOfLibrary(t, g, me, "Enlightened Tutor", "Instant", effectstest.EnlightenedTutorOracle, needle)

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
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Not creatures: must not be offered.
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bolt", TypeLine: "Instant"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Craterhoof Behemoth", TypeLine: "Creature — Beast",
	})

	top := effectstest.TutorTopOfLibrary(t, g, me, "Worldly Tutor", "Instant", effectstest.WorldlyTutorOracle, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored creature", top.Name)
	}
	if !top.IsKnownTo(opp.ID) {
		t.Error("an opponent does not know the revealed card; Worldly Tutor says reveal it")
	}
	if me.Hand.Contains(needle) {
		t.Error("the tutored card went to hand; this tutor puts it on top")
	}
}

func TestMysticalTutorFindsAnInstantOrSorcery(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	// Not instants or sorceries: must not be offered.
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Bear", TypeLine: "Creature — Bear"})
	effectstest.PushLibraryCardForTest(me, game.Card{Name: "Sol Ring", TypeLine: "Artifact"})
	needle := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Counterspell", TypeLine: "Instant",
	})

	top := effectstest.TutorTopOfLibrary(t, g, me, "Mystical Tutor", "Instant", effectstest.MysticalTutorOracle, needle)
	if top.InstanceID != needle {
		t.Errorf("top of library is %q, want the tutored instant", top.Name)
	}
	if !top.IsKnownTo(opp.ID) {
		t.Error("an opponent does not know the revealed card; Mystical Tutor says reveal it")
	}
	if me.Hand.Contains(needle) {
		t.Error("the tutored card went to hand; this tutor puts it on top")
	}
}

// TestImperialSealIsPrivateAndCostsLife — the one card in the cycle
// with no "reveal it". Nobody but the controller learns the top card,
// and the 2 life is what buys that.
func TestImperialSealIsPrivateAndCostsLife(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Anything At All", TypeLine: "Sorcery"})
	lifeBefore := me.Life

	top := effectstest.TutorTopOfLibrary(t, g, me, "Imperial Seal", "Sorcery", effectstest.ImperialSealOracle, needle)

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	for i := 0; i < 30; i++ {
		effectstest.PushLibraryCardForTest(me, game.Card{Name: "Chaff", TypeLine: "Sorcery"})
	}
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "The Needle", TypeLine: "Instant"})

	top := effectstest.TutorTopOfLibrary(t, g, me, "Mystical Tutor", "Instant", effectstest.MysticalTutorOracle, needle)
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	// Two lands, so mana value 2 is findable and 5 is not.
	effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	effectstest.PushCatalogPermanent(g, me.ID, "Island", "Basic Land — Island", "", false)
	cheap := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Two Drop", TypeLine: "Creature — Human", ManaCost: "{1}{U}",
	})
	expensive := effectstest.PushLibraryCardForTest(me, game.Card{
		Name: "Five Drop", TypeLine: "Creature — Giant", ManaCost: "{4}{U}",
	})

	effectstest.CastCatalogSpell(t, g, "Beseech the Queen", "Sorcery", effectstest.BeseechOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
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

	effectstest.AnswerSearchByID(t, g, me.ID, cheap)
	if !me.Hand.Contains(cheap) {
		t.Error("the tutored card is not in hand")
	}
}

// TestTutorSearchTypeRestrictions pins the card type each tutor's
// search allows: two legal cards (so the engine must prompt) and one
// illegal one; only the legal ones are offered.
func TestTutorSearchTypeRestrictions(t *testing.T) {
	cases := []struct {
		name, typeLine, oracle string
		good, bad              string
	}{
		{"Worldly Tutor", "Instant", effectstest.WorldlyTutorOracle, "Creature — Bear", "Instant"},
		{"Mystical Tutor", "Instant", effectstest.MysticalTutorOracle, "Sorcery", "Creature — Bear"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			g := effectstest.NewCatalogGame(t)
			me := g.Seats[0]
			g1 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "G1", TypeLine: tc.good})
			g2 := effectstest.PushLibraryCardForTest(me, game.Card{Name: "G2", TypeLine: tc.good})
			b := effectstest.PushLibraryCardForTest(me, game.Card{Name: "B", TypeLine: tc.bad})
			effectstest.CastCatalogSpell(t, g, tc.name, tc.typeLine, tc.oracle, nil)
			effectstest.PassPriorityAroundTable(t, g)
			ch := effectstest.SearchChoiceFor(g, me.ID)
			if ch == nil {
				t.Fatal("no search prompt")
			}
			offered := map[uuid.UUID]bool{}
			for _, id := range ch.SearchCards {
				offered[id] = true
			}
			if !offered[g1] || !offered[g2] || offered[b] {
				t.Errorf("offered %v, want only the two matching cards", offered)
			}
		})
	}
}
