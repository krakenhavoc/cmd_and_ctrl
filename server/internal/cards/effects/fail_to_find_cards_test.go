package effects

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// fail_to_find_cards_test.go — #3053 / #3081 through real catalog
// cards. The engine rule (CR 701.23b/d) is pinned in the game package;
// this file pins that the cards inherit it without a per-card flag:
//
//   - a basic-land fetch against exactly one basic still asks, and
//     "find nothing" is honoured with the shuffle intact;
//   - the "up to N" cards of #3081 may take fewer than N (and 0);
//   - Demonic Tutor and Intuition — no stated quality — must find.

const demonicTutorOracle = "82004860-e589-4e38-8d61-8c0210e4ea39"

func searchEventCount(g *game.Game) int {
	n := 0
	g.ReadSnapshot(func() {
		for _, ev := range g.Events {
			if ev.Kind == game.EventSearchLibrary {
				n++
			}
		}
	})
	return n
}

// libraryKnownTo counts library cards the seat still knows the identity
// of: zero after a shuffle (the shuffle un-knows the zone).
func libraryKnownTo(p *game.Player, who uuid.UUID) int {
	n := 0
	for _, c := range p.Library.Cards {
		if c.KnownBy[who] {
			n++
		}
	}
	return n
}

// The #3053 headline: Rampant Growth into a library holding exactly one
// basic. It used to put the Forest onto the battlefield without asking.
func TestRampantGrowthWithOneBasicOffersFailToFind(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me,
		searchTestLand("Shock", "Instant"),
		searchTestLand("Forest", "Basic Land — Forest"),
		searchTestLand("Bolt", "Instant"),
	)
	forest := ids[1]

	castCatalogSpell(t, g, "Rampant Growth", "Sorcery", rampantGrowthOracle, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("one matching basic: the search did not ask (CR 701.23b)")
	}
	if c.SearchMin != 0 || c.SearchMax != 1 || len(c.SearchCards) != 1 {
		t.Fatalf("prompt min/max/candidates = %d/%d/%d, want 0/1/1", c.SearchMin, c.SearchMax, len(c.SearchCards))
	}
	if g.Battlefield.Contains(forest) {
		t.Fatal("the Forest was fetched before the searcher answered")
	}

	answerSearchFailToFind(t, g, me.ID)

	if g.Battlefield.Contains(forest) || !me.Library.Contains(forest) {
		t.Error("finding nothing still moved the Forest")
	}
	if searchEventCount(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: it was still a search", searchEventCount(g))
	}
	if libraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled after finding nothing")
	}
}

func TestRampantGrowthWithOneBasicCanStillTakeIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))

	castCatalogSpell(t, g, "Rampant Growth", "Sorcery", rampantGrowthOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0])

	if card, ok := battlefieldCard(g, ids[0]); !ok || !card.Tapped {
		t.Errorf("the Forest should be on the battlefield tapped: %+v ok=%v", card, ok)
	}
}

// #3081: "up to two basic land cards" with exactly two in the library
// was taken whole. Take both, one, or none.
func TestExplosiveVegetationMayTakeFewerThanTwo(t *testing.T) {
	for _, take := range []int{0, 1, 2} {
		g := newCatalogGame(t)
		me := g.Seats[0]
		ids := seedBasicsInLibrary(me, 2)

		castCatalogSpell(t, g, "Explosive Vegetation", "Sorcery", explosiveVegetationOracle, nil)
		passPriorityAroundTable(t, g)
		c := searchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("take %d: exactly two basics and no prompt: the second could not be declined", take)
		}
		if c.SearchMin != 0 || c.SearchMax != 2 {
			t.Fatalf("take %d: min/max = %d/%d, want 0/2", take, c.SearchMin, c.SearchMax)
		}
		answerSearchByID(t, g, me.ID, ids[:take]...)

		onField := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				onField++
			}
		}
		if onField != take {
			t.Errorf("take %d: %d basics reached the battlefield", take, onField)
		}
		if searchEventCount(g) != 1 {
			t.Errorf("take %d: EventSearchLibrary fired %d times, want 1", take, searchEventCount(g))
		}
		if libraryKnownTo(me, me.ID) != 0 {
			t.Errorf("take %d: the library was not shuffled", take)
		}
	}
}

func TestSkyshroudClaimMayFindOnlyOneForest(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedBasicsInLibrary(me, 2)

	castCatalogSpell(t, g, "Skyshroud Claim", "Sorcery", skyshroudClaimOracle, nil)
	passPriorityAroundTable(t, g)
	answerSearchByID(t, g, me.ID, ids[0])

	if !g.Battlefield.Contains(ids[0]) {
		t.Error("the Forest that was taken is not on the battlefield")
	}
	if g.Battlefield.Contains(ids[1]) {
		t.Error("the Forest that was declined was fetched anyway")
	}
}

// Gifts Ungiven: "up to four cards with different names" — and zero is a
// legal find, after which the library is shuffled and nothing is offered
// to the opponent.
func TestGiftsUngivenMayFindNothing(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := seedTypedLibrary(me, "Artifact", "Gift One", "Gift Two")

	castCatalogSpell(t, g, "Gifts Ungiven", "Instant", giftsUngivenOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	c := searchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMin != 0 || c.SearchMax != 4 {
		t.Fatalf("min/max = %d/%d, want 0/4", c.SearchMin, c.SearchMax)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing: %v", err)
	}
	if latestRevealPickFor(g, opp.ID) != nil {
		t.Error("the opponent was asked to choose among nothing")
	}
	for _, id := range seeded {
		if !me.Library.Contains(id) {
			t.Errorf("%s left the library", id)
		}
	}
}

// Demonic Tutor: "a card" names no quality, so the caster must find one.
func TestDemonicTutorMustFindACard(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	needle := pushLibraryCardForTest(me, game.Card{Name: "Needle"})

	castCatalogSpell(t, g, "Demonic Tutor", "Sorcery", demonicTutorOracle, nil)
	passPriorityAroundTable(t, g)

	c := searchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMin != 1 || c.SearchMax != 1 {
		t.Fatalf("min/max = %d/%d, want 1/1: Demonic Tutor has no fail to find", c.SearchMin, c.SearchMax)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("failing to find Demonic Tutor: err = %v, want ErrInvalidParam", err)
	}
	if searchChoiceFor(g, me.ID) == nil {
		t.Fatal("the refused answer dequeued the prompt")
	}
	answerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the card was not tutored")
	}
}

func TestDemonicTutorWithOneCardLeftTakesIt(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	ids := seedSearchLibrary(me, searchTestLand("Lonely", "Creature — Bear"))

	castCatalogSpell(t, g, "Demonic Tutor", "Sorcery", demonicTutorOracle, nil)
	passPriorityAroundTable(t, g)

	if searchChoiceFor(g, me.ID) != nil {
		t.Fatal("a forced pick queued a prompt")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the last card in the library was not found")
	}
}

// Intuition: "three cards" — three, not "up to".
func TestIntuitionMustFindThreeCards(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := seedTypedLibrary(me, "Artifact", "One", "Two", "Three", "Four")

	castCatalogSpell(t, g, "Intuition", "Instant", intuitionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	passPriorityAroundTable(t, g)

	c := searchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMin != 3 || c.SearchMax != 3 {
		t.Fatalf("min/max = %d/%d, want 3/3", c.SearchMin, c.SearchMax)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, seeded[:2]); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("two of three: err = %v, want ErrInvalidParam", err)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, seeded[:3]); err != nil {
		t.Fatalf("three of three: %v", err)
	}
}

// A "you MAY search" card keeps asking, and a card that does not say
// "may" does too, now with no flag. Pins the two together so the
// distinction in the docs is also in the behaviour.
func TestOptionalAndPlainSearchesBothOfferFailToFind(t *testing.T) {
	for _, optional := range []bool{false, true} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		seedSearchLibrary(me, searchTestLand("Forest", "Basic Land — Forest"))
		g.WithWriteLock(func() {
			_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player: me.ID, Pred: IsBasicLand, Dest: game.ZoneBattlefield,
				Limit: 1, Optional: optional, Shuffle: true,
			})
		})
		c := searchChoiceFor(g, me.ID)
		if c == nil || c.SearchMin != 0 {
			t.Errorf("Optional=%v: prompt = %+v, want one with a zero floor", optional, c)
		}
	}
}
