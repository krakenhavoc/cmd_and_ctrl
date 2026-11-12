package effects

import (
	"slices"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// urza_assembles_the_titans_test.go — Urza Assembles the Titans (#5665)
// and the one-turn "twice this turn" loyalty limit its chapter III
// needs (Game.AllowLoyaltyTwiceThisTurnForEffect).

const urzaAssemblesOracle = "0d406496-e1c3-4e1c-976c-618d5ad67d0b"

// urzaAssemblesAt casts the Saga from the active seat and starts it on
// `chapter`.
func urzaAssemblesAt(t *testing.T, g *game.Game, chapter int) uuid.UUID {
	t.Helper()
	id := castReadAheadSaga(t, g, "Urza Assembles the Titans", urzaAssemblesOracle)
	chooseChapter(t, g, chapter)
	passPriorityAroundTable(t, g)
	return id
}

func urzaTestWalkerCard(name, cost string) game.Card {
	return game.Card{Name: name, TypeLine: "Legendary Planeswalker — Test", ManaCost: cost, StartingLoyalty: 5}
}

func urzaTopOfLibrary(p *game.Player) uuid.UUID {
	if p.Library.Size() == 0 {
		return uuid.Nil
	}
	return p.Library.Cards[p.Library.Size()-1].InstanceID
}

func TestUrzaAssemblesTheTitansIsRegisteredFullWithNoCaveats(t *testing.T) {
	spec, ok := Lookup(urzaAssemblesOracle)
	if !ok {
		t.Fatal("not registered")
	}
	if spec.Completeness != CompletenessFull || len(spec.Caveats) != 0 {
		t.Errorf("%v with caveats %v, want full with none", spec.Completeness, spec.Caveats)
	}
}

// Chapter I scries 4, then asks about the reveal once the scry is
// done; a revealed planeswalker card goes to the hand. The scry prompt
// and the "you may" are both restore points.
func TestUrzaAssemblesChapterOneScriesThenRevealsAPlaneswalkerIntoHand(t *testing.T) {
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	ids := setLibraryTopFirst(me,
		urzaTestWalkerCard("Top Walker", "{3}{U}"), libraryCreature("Second"),
		libraryCreature("Third"), libraryCreature("Fourth"), libraryCreature("Fifth"))

	saga := urzaAssemblesAt(t, g, 1)
	scry := scryChoiceFor(g, me.ID)
	if scry == nil {
		t.Fatal("chapter I did not scry")
	}
	if len(scry.ScryCards) != 4 {
		t.Fatalf("scry looks at %d cards, want 4", len(scry.ScryCards))
	}
	if confirmChoiceFor(g, me.ID) != nil {
		t.Fatal("the reveal was asked before the scry was done")
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[seat]
	answerScryKeepAll(t, g, me.ID)
	g, ask := restoreAtConfirm(t, g, me.ID, urzaAssemblesReveal.Key())
	me = g.Seats[seat]
	hand := len(me.Hand.Cards)
	if err := g.ResolveConfirm(ask.ID, me.ID, true); err != nil {
		t.Fatalf("ResolveConfirm: %v", err)
	}
	passPriorityAroundTable(t, g)

	if !inHandOf(me, ids[0]) {
		t.Error("the revealed planeswalker card was not put into the hand")
	}
	if got := len(me.Hand.Cards); got != hand+1 {
		t.Errorf("hand = %d, want %d", got, hand+1)
	}
	if urzaTopOfLibrary(me) != ids[1] {
		t.Error("the library below the revealed card moved")
	}
	if !g.Battlefield.Contains(saga) {
		t.Error("the Saga left after chapter I")
	}
}

// A revealed card that is not a planeswalker card stays on top, and a
// "no" reveals nothing.
func TestUrzaAssemblesChapterOneLeavesAnyOtherCardOnTop(t *testing.T) {
	for _, accept := range []bool{true, false} {
		g := newCatalogGame(t)
		me := g.Seats[g.Turn.ActiveSeat]
		ids := setLibraryTopFirst(me,
			libraryCreature("Top"), urzaTestWalkerCard("Buried Walker", "{3}{U}"), libraryCreature("Third"))

		urzaAssemblesAt(t, g, 1)
		answerScryKeepAll(t, g, me.ID)
		hand := len(me.Hand.Cards)
		answerConfirmFor(t, g, me, accept)
		passPriorityAroundTable(t, g)

		if got := len(me.Hand.Cards); got != hand {
			t.Errorf("accept=%v: hand = %d, want %d", accept, got, hand)
		}
		if urzaTopOfLibrary(me) != ids[0] {
			t.Errorf("accept=%v: the top card moved", accept)
		}
		revealed := false
		for _, ev := range g.Events {
			if ev.Kind == game.EventRevealCards {
				revealed = true
			}
		}
		if revealed != accept {
			t.Errorf("accept=%v: revealed = %v", accept, revealed)
		}
	}
}

// With an empty library there is nothing to reveal, so nothing is asked.
func TestUrzaAssemblesChapterOneAsksNothingOverAnEmptyLibrary(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	setLibraryTopFirst(me)

	urzaAssemblesAt(t, g, 1)
	if scryChoiceFor(g, me.ID) != nil || confirmChoiceFor(g, me.ID) != nil {
		t.Errorf("an empty library was asked about: %+v", g.PendingChoices)
	}
}

// Chapter II offers only planeswalker cards with mana value 6 or less,
// and the pick enters the battlefield with its loyalty.
func TestUrzaAssemblesChapterTwoPutsAPlaneswalkerOfManaValueSixOrLess(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	six := putInHand(me, urzaTestWalkerCard("Six", "{4}{U}{U}"))
	seven := putInHand(me, urzaTestWalkerCard("Seven", "{5}{U}{U}"))
	creature := pushHandCardWithManaCost(me, "Bear", "Creature — Bear", "{1}{G}")

	id := urzaAssemblesAt(t, g, 2)
	if got := chaptersFired(g, id); !slices.Equal(got, []int{2}) {
		t.Fatalf("chapters fired = %v, want only [2]", got)
	}
	pick := chooseCardsChoiceFor(g, me.ID)
	if pick == nil {
		t.Fatal("chapter II did not offer the put")
	}
	if pick.ChooseMin != 0 || pick.ChooseMax != 1 {
		t.Errorf("bounds %d..%d, want 0..1", pick.ChooseMin, pick.ChooseMax)
	}
	if !slices.Equal(pick.ChooseCards, []uuid.UUID{six}) {
		t.Errorf("candidates = %v, want only the mana value 6 planeswalker (not %s or %s)", pick.ChooseCards, seven, creature)
	}
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{seven}); err == nil {
		t.Fatal("a mana value 7 planeswalker was accepted")
	}
	answerChooseCards(t, g, me.ID, six)
	passPriorityAroundTable(t, g)

	if !g.Battlefield.Contains(six) {
		t.Fatal("the planeswalker did not enter the battlefield")
	}
	if got := loyaltyOf(g, six); got != 5 {
		t.Errorf("loyalty = %d, want 5", got)
	}
	for _, ev := range g.Events {
		if ev.Kind == game.EventCast && ev.CardID == six {
			t.Error("the planeswalker was cast, not put")
		}
	}
}

// Chapter II is "you may": declining puts nothing.
func TestUrzaAssemblesChapterTwoMayBeDeclined(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	six := putInHand(me, urzaTestWalkerCard("Six", "{4}{U}{U}"))

	urzaAssemblesAt(t, g, 2)
	answerChooseCards(t, g, me.ID)
	passPriorityAroundTable(t, g)
	if g.Battlefield.Contains(six) || !inHandOf(me, six) {
		t.Error("a declined put moved the card")
	}
}

// Chapter III: the limit is two for the rest of the turn, counting an
// activation made before it resolved, for a planeswalker that arrives
// afterwards too; the enumerator and the wire agree; the effect is a
// restore point's data and ends with the turn.
func TestUrzaAssemblesChapterThreeMakesTheLoyaltyLimitTwoThisTurn(t *testing.T) {
	g, me, _ := spendTable(t)
	seat := g.Turn.ActiveSeat
	c := guffbChandra(g, me.ID)
	chandraPlusOne(t, g, me, c)
	if loyaltyRowsOffered(g, me, c) != 0 {
		t.Fatal("setup: a spent walker is still offered")
	}

	saga := urzaAssemblesAt(t, g, 3)
	if got := chaptersFired(g, saga); !slices.Equal(got, []int{3}) {
		t.Fatalf("chapters fired = %v, want only [3]", got)
	}
	if g.Battlefield.Contains(saga) {
		t.Error("the Saga was not sacrificed after its final chapter")
	}
	if loyaltyRowsOffered(g, me, c) == 0 {
		t.Error("the enumerator does not offer the second activation")
	}
	if walkerGreyed(t, g, me, c) {
		t.Error("the wire still greys the walker after chapter III resolved")
	}

	g = restoreRoundTrip(t, g, true)
	me = g.Seats[seat]
	if loyaltyRowsOffered(g, me, c) == 0 {
		t.Fatal("the restored game lost the second activation")
	}
	// A walker that arrives after the chapter resolved is covered.
	late := guffbWalker(g, me.ID, "Second Walker", "Planeswalker — Chandra", guffbChandraOracle, 4)

	chandraPlusOne(t, g, me, c)
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a third activation was allowed")
	}
	if loyaltyRowsOffered(g, me, c) != 0 || !walkerGreyed(t, g, me, c) {
		t.Error("the walker is still offered after its two activations")
	}
	if got := loyaltyOf(g, c); got != 6 {
		t.Errorf("loyalty %d, want 6 after two +1s", got)
	}
	chandraPlusOne(t, g, me, late)
	chandraPlusOne(t, g, me, late)
	if err := g.ActivateCatalogAbility(me.ID, late, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the later walker made a third activation")
	}

	// "This turn": on the player's next turn the limit is one again.
	advanceToPrecombatMainOf(t, g, seat)
	chandraPlusOne(t, g, me, c)
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("the effect outlived its turn")
	}
}

// The chapter gives nothing to another player's planeswalkers and no
// second activation of a noncreature permanent's loyalty ability that
// is not a planeswalker's is implied: without the chapter the limit is
// one.
func TestUrzaAssemblesChapterThreeIsNeededForTheSecondActivation(t *testing.T) {
	g, me, _ := spendTable(t)
	c := guffbChandra(g, me.ID)
	chandraPlusOne(t, g, me, c)

	// Chapter II resolves instead: no permission.
	urzaAssemblesAt(t, g, 2)
	if err := g.ActivateCatalogAbility(me.ID, c, guffbChandraAddRR, game.ActivateAbilityParams{}); err == nil {
		t.Fatal("a second activation was allowed without chapter III")
	}
}
