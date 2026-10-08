package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// ADR 0113's amendment of 2026-10-08 (#2108): a maximum-hand-size
// change a spell makes, for a duration. Inspired Idea and Enter the
// Infinite are the proof cards.

const (
	mhInspiredIdeaOracle     = "5f1386d5-b802-46b2-9803-275621ebda80"
	mhEnterTheInfiniteOracle = "2adbb56a-45e9-4fbe-b586-3488ef8014a3"
)

func castInspiredIdea(t *testing.T, g *game.Game, cleave bool) {
	t.Helper()
	active := g.Seats[g.Turn.ActiveSeat]
	id := uuid.New()
	active.Hand.PushTop(game.Card{
		InstanceID: id, Name: "Inspired Idea", TypeLine: "Sorcery",
		OracleID: mhInspiredIdeaOracle, Owner: active.ID, Controller: active.ID,
	})
	for g.Turn.Step != game.StepPrecombatMain {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	params := game.CastSpellParams{}
	if cleave {
		params.AlternativeCost = "cleave"
	}
	if err := g.CastSpell(active.ID, id, params); err != nil {
		t.Fatalf("cast Inspired Idea: %v", err)
	}
	passPriorityAroundTable(t, g)
}

func TestInspiredIdeaReducesTheMaximumByThreeForTheRestOfTheGame(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castInspiredIdea(t, g, false)
	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand = %d, want %d (drew three)", got, before+3)
	}
	if got := mhMax(g, me); got != 4 {
		t.Errorf("max hand size = %d, want 4", got)
	}
	for _, p := range g.Seats {
		if p.ID != me.ID && mhMax(g, p) != 7 {
			t.Errorf("an opponent's maximum changed")
		}
	}
	// A second one reduces by three again; the engine does not clamp
	// the running value, the result is max(0, 1) = 1.
	castInspiredIdea(t, g, false)
	if got := mhMax(g, me); got != 1 {
		t.Errorf("two Inspired Ideas: max = %d, want 1", got)
	}
	// "For the rest of the game" survives the turn passing.
	for seat := g.Turn.ActiveSeat; g.Turn.ActiveSeat == seat; {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if got := mhMax(g, me); got != 1 {
		t.Errorf("after the turn passes: max = %d, want 1", got)
	}
}

func TestCleavedInspiredIdeaOnlyDraws(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	before := me.Hand.Size()
	castInspiredIdea(t, g, true)
	if got := me.Hand.Size(); got != before+3 {
		t.Fatalf("hand = %d, want %d", got, before+3)
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("cleaved: max = %d, want 7 (the bracketed words are removed)", got)
	}
}

// CR 613.11: the reduction sorts by timestamp with the permanents'
// statics. Inspired Idea, then Null Profusion, is two; Null Profusion,
// then Inspired Idea, is two minus three, clamped to zero.
func TestInspiredIdeaFoldsInTimestampOrder(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	castInspiredIdea(t, g, false)
	mhNullProfusion(g, me.ID)
	if got := mhMax(g, me); got != 2 {
		t.Errorf("Idea then Null Profusion: max = %d, want 2", got)
	}

	g2 := newCatalogGame(t)
	me2 := g2.Seats[g2.Turn.ActiveSeat]
	mhNullProfusion(g2, me2.ID)
	castInspiredIdea(t, g2, false)
	if got := mhMax(g2, me2); got != 0 {
		t.Errorf("Null Profusion then Idea: max = %d, want 0", got)
	}
}

func TestEnterTheInfiniteDrawsTheLibraryPutsOneBackAndLiftsTheCapUntilYourNextTurn(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	seat := g.Turn.ActiveSeat
	me := g.Seats[seat]
	library := me.Library.Size()
	handBefore := me.Hand.Size()
	castCatalogSpell(t, g, "Enter the Infinite", "Sorcery", mhEnterTheInfiniteOracle, nil)
	passPriorityAroundTable(t, g)

	if me.Library.Size() != 0 {
		t.Fatalf("library = %d, want it all drawn (was %d)", me.Library.Size(), library)
	}
	pick := latestChooseCardsFor(g, me.ID)
	if pick == nil {
		t.Fatal("no put-back prompt")
	}
	if pick.ChooseMin != 1 || pick.ChooseMax != 1 {
		t.Errorf("put-back bounds = %d..%d, want exactly one", pick.ChooseMin, pick.ChooseMax)
	}
	card := me.Hand.Cards[0].InstanceID
	if err := g.ResolveChooseCards(pick.ID, me.ID, []uuid.UUID{card}); err != nil {
		t.Fatalf("ResolveChooseCards: %v", err)
	}
	if got := me.Hand.Size(); got != handBefore+library-1 {
		t.Errorf("hand = %d, want %d", got, handBefore+library-1)
	}
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max = %d, want no maximum", got)
	}

	// It lasts through the other players' turns...
	for g.Turn.ActiveSeat == seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("on another player's turn: max = %d, want no maximum", got)
	}
	// ...and ends as the caster's next turn begins.
	for g.Turn.ActiveSeat != seat {
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
	if got := mhMax(g, me); got != 7 {
		t.Errorf("on the caster's next turn: max = %d, want 7", got)
	}
}

// The library may be empty: the draw takes nothing and the put-back
// still has a card to place only if the hand has one.
func TestEnterTheInfiniteWithAnEmptyLibraryStillGrantsTheCap(t *testing.T) {
	mhTickingClock(t)
	g := newCatalogGame(t)
	me := g.Seats[g.Turn.ActiveSeat]
	me.Library.Cards = nil
	castCatalogSpell(t, g, "Enter the Infinite", "Sorcery", mhEnterTheInfiniteOracle, nil)
	passPriorityAroundTable(t, g)
	if got := mhMax(g, me); got != game.NoMaxHandSize {
		t.Errorf("max = %d, want no maximum", got)
	}
}

func TestGrantHandSizeRefusesTheTwoMeaninglessDeclarations(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		if err := g.GrantHandSizeForEffect(me.ID, game.HandSizeModify, 0, "x", uuid.Nil, game.IndefiniteDuration()); err == nil {
			t.Error("a modify by zero must be refused")
		}
		if err := g.GrantHandSizeForEffect(me.ID, game.HandSizeSet, -1, "x", uuid.Nil, game.IndefiniteDuration()); err == nil {
			t.Error("a set below zero must be refused")
		}
		if err := g.GrantHandSizeForEffect(uuid.New(), game.HandSizeSet, 1, "x", uuid.Nil, game.IndefiniteDuration()); err == nil {
			t.Error("an unknown player must be refused")
		}
	})
}
