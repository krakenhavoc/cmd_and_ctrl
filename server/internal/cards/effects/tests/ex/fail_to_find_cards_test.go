package testsex

import (
	"errors"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects/effectstest"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The #3053 headline: Rampant Growth into a library holding exactly one
// basic. It used to put the Forest onto the battlefield without asking.
func TestRampantGrowthWithOneBasicOffersFailToFind(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me,
		effectstest.SearchTestLand("Shock", "Instant"),
		effectstest.SearchTestLand("Forest", "Basic Land — Forest"),
		effectstest.SearchTestLand("Bolt", "Instant"),
	)
	forest := ids[1]

	effectstest.CastCatalogSpell(t, g, "Rampant Growth", "Sorcery", effectstest.RampantGrowthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("one matching basic: the search did not ask (CR 701.23b)")
	}
	if c.SearchMin != 0 || c.SearchMax != 1 || len(c.SearchCards) != 1 {
		t.Fatalf("prompt min/max/candidates = %d/%d/%d, want 0/1/1", c.SearchMin, c.SearchMax, len(c.SearchCards))
	}
	if g.Battlefield.Contains(forest) {
		t.Fatal("the Forest was fetched before the searcher answered")
	}

	effectstest.AnswerSearchFailToFind(t, g, me.ID)

	if g.Battlefield.Contains(forest) || !me.Library.Contains(forest) {
		t.Error("finding nothing still moved the Forest")
	}
	if effectstest.SearchEventCount(g) != 1 {
		t.Errorf("EventSearchLibrary fired %d times, want 1: it was still a search", effectstest.SearchEventCount(g))
	}
	if effectstest.LibraryKnownTo(me, me.ID) != 0 {
		t.Error("the library was not shuffled after finding nothing")
	}
}

func TestRampantGrowthWithOneBasicCanStillTakeIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))

	effectstest.CastCatalogSpell(t, g, "Rampant Growth", "Sorcery", effectstest.RampantGrowthOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])

	if card, ok := effectstest.BattlefieldCard(g, ids[0]); !ok || !card.Tapped {
		t.Errorf("the Forest should be on the battlefield tapped: %+v ok=%v", card, ok)
	}
}

// #3081: "up to two basic land cards" with exactly two in the library
// was taken whole. Take both, one, or none.
func TestExplosiveVegetationMayTakeFewerThanTwo(t *testing.T) {
	for _, take := range []int{0, 1, 2} {
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[0]
		ids := effectstest.SeedBasicsInLibrary(me, 2)

		effectstest.CastCatalogSpell(t, g, "Explosive Vegetation", "Sorcery", effectstest.ExplosiveVegetationOracle, nil)
		effectstest.PassPriorityAroundTable(t, g)
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil {
			t.Fatalf("take %d: exactly two basics and no prompt: the second could not be declined", take)
		}
		if c.SearchMin != 0 || c.SearchMax != 2 {
			t.Fatalf("take %d: min/max = %d/%d, want 0/2", take, c.SearchMin, c.SearchMax)
		}
		effectstest.AnswerSearchByID(t, g, me.ID, ids[:take]...)

		onField := 0
		for _, id := range ids {
			if g.Battlefield.Contains(id) {
				onField++
			}
		}
		if onField != take {
			t.Errorf("take %d: %d basics reached the battlefield", take, onField)
		}
		if effectstest.SearchEventCount(g) != 1 {
			t.Errorf("take %d: EventSearchLibrary fired %d times, want 1", take, effectstest.SearchEventCount(g))
		}
		if effectstest.LibraryKnownTo(me, me.ID) != 0 {
			t.Errorf("take %d: the library was not shuffled", take)
		}
	}
}

func TestSkyshroudClaimMayFindOnlyOneForest(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedBasicsInLibrary(me, 2)

	effectstest.CastCatalogSpell(t, g, "Skyshroud Claim", "Sorcery", effectstest.SkyshroudClaimOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)
	effectstest.AnswerSearchByID(t, g, me.ID, ids[0])

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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := effectstest.SeedTypedLibrary(me, "Artifact", "Gift One", "Gift Two")

	effectstest.CastCatalogSpell(t, g, "Gifts Ungiven", "Instant", effectstest.GiftsUngivenOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchPromptFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMin != 0 || c.SearchMax != 4 {
		t.Fatalf("min/max = %d/%d, want 0/4", c.SearchMin, c.SearchMax)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); err != nil {
		t.Fatalf("finding nothing: %v", err)
	}
	if effectstest.LatestRevealPickFor(g, opp.ID) != nil {
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
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	needle := effectstest.PushLibraryCardForTest(me, game.Card{Name: "Needle"})

	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchChoiceFor(g, me.ID)
	if c == nil {
		t.Fatal("no search prompt")
	}
	if c.SearchMin != 1 || c.SearchMax != 1 {
		t.Fatalf("min/max = %d/%d, want 1/1: Demonic Tutor has no fail to find", c.SearchMin, c.SearchMax)
	}
	if err := g.ResolveSearchLibrary(c.ID, me.ID, nil); !errors.Is(err, game.ErrInvalidParam) {
		t.Fatalf("failing to find Demonic Tutor: err = %v, want ErrInvalidParam", err)
	}
	if effectstest.SearchChoiceFor(g, me.ID) == nil {
		t.Fatal("the refused answer dequeued the prompt")
	}
	effectstest.AnswerSearchByID(t, g, me.ID, needle)
	if !me.Hand.Contains(needle) {
		t.Error("the card was not tutored")
	}
}

func TestDemonicTutorWithOneCardLeftTakesIt(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[0]
	ids := effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Lonely", "Creature — Bear"))

	effectstest.CastCatalogSpell(t, g, "Demonic Tutor", "Sorcery", effectstest.DemonicTutorOracle, nil)
	effectstest.PassPriorityAroundTable(t, g)

	if effectstest.SearchChoiceFor(g, me.ID) != nil {
		t.Fatal("a forced pick queued a prompt")
	}
	if !me.Hand.Contains(ids[0]) {
		t.Error("the last card in the library was not found")
	}
}

// Intuition: "three cards" — three, not "up to".
func TestIntuitionMustFindThreeCards(t *testing.T) {
	g := effectstest.NewCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	opp := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	seeded := effectstest.SeedTypedLibrary(me, "Artifact", "One", "Two", "Three", "Four")

	effectstest.CastCatalogSpell(t, g, "Intuition", "Instant", effectstest.IntuitionOracle,
		[]game.TargetRef{{Kind: game.TargetPlayer, ID: opp.ID}})
	effectstest.PassPriorityAroundTable(t, g)

	c := effectstest.SearchPromptFor(g, me.ID)
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
		g := effectstest.NewCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		effectstest.SeedSearchLibrary(me, effectstest.SearchTestLand("Forest", "Basic Land — Forest"))
		g.WithWriteLock(func() {
			_ = g.SearchLibraryThenForEffect(game.SearchLibrarySpec{
				Player: me.ID, Pred: effects.IsBasicLand, Dest: game.ZoneBattlefield,
				Limit: 1, Optional: optional, Shuffle: true,
			})
		})
		c := effectstest.SearchChoiceFor(g, me.ID)
		if c == nil || c.SearchMin != 0 {
			t.Errorf("Optional=%v: prompt = %+v, want one with a zero floor", optional, c)
		}
	}
}
