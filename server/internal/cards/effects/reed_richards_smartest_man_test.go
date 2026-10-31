package effects

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// reed_richards_smartest_man_test.go — #4455: "The first time you would
// draw a card each turn except the first card you draw during each of your
// draw steps, you draw four cards instead." Driven through the engine's
// draw path (DrawNForEffect and the turn-based draw), never by hand.

const reedRichardsOracle = "51f55021-04e0-49bf-8f1f-de2a578c95fc"

func reedOnTable(t *testing.T) (*game.Game, *game.Player) {
	t.Helper()
	g := newCatalogGame(t)
	me := g.Seats[0]
	pushCatalogPermanent(g, me.ID, "Reed Richards, Smartest Man", "Legendary Creature — Human Scientist Hero", reedRichardsOracle, false)
	return g, me
}

func TestReedRichardsHasNoMaximumHandSize(t *testing.T) {
	g, me := reedOnTable(t)
	if got := g.EffectiveMaxHandSizeLocked(me); got != game.NoMaxHandSize {
		t.Errorf("max hand size is %d with Reed out, want the no-limit sentinel %d", got, game.NoMaxHandSize)
	}
	opp := g.Seats[1]
	if got := g.EffectiveMaxHandSizeLocked(opp); got == game.NoMaxHandSize {
		t.Error("Reed gave an opponent no maximum hand size")
	}
}

func TestReedRichardsFirstDrawOfTheTurnIsFour(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("the first draw of the turn drew %d cards, want 4", got)
	}
}

func TestReedRichardsOnlyTheFirstTimeEachTurn(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Fatalf("the first draw drew %d cards, want 4", got)
	}
	if got := drawFor(t, g, me, 1); got != 1 {
		t.Errorf("a second draw the same turn drew %d cards, want 1 (it is the first time only)", got)
	}
	if got := drawFor(t, g, me, 2); got != 2 {
		t.Errorf("draw two after that drew %d cards, want 2", got)
	}
}

// "Draw three" is three individual draws (CR 121.2): the first is the
// first time and becomes four, the other two are not.
func TestReedRichardsReplacesOnlyTheFirstOfThreeDraws(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	if got := drawFor(t, g, me, 3); got != 6 {
		t.Errorf("draw three drew %d cards, want 4 + 1 + 1 = 6", got)
	}
}

// The four cards are not themselves replaced by Reed again (CR 614.5).
func TestReedRichardsFourCardsAreNotReplacedAgain(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	before := len(g.DrawnThisTurn[me.ID])
	drawFor(t, g, me, 1)
	if got := len(g.DrawnThisTurn[me.ID]) - before; got != 4 {
		t.Errorf("%d cards counted as drawn, want 4", got)
	}
}

// The first card drawn in your own draw step is exempt, and it does not
// use up the turn's one replacement: a later draw this turn is four.
func TestReedRichardsDrawStepCardIsExemptAndSavesTheReplacement(t *testing.T) {
	g := newCatalogGame(t)
	me := g.Seats[1]
	pushCatalogPermanent(g, me.ID, "Reed Richards, Smartest Man", "Legendary Creature — Human Scientist Hero", reedRichardsOracle, false)
	before := me.Hand.Size()
	advanceToDrawStepOfSeat(t, g, 1)
	if got := me.Hand.Size() - before; got != 1 {
		t.Fatalf("the turn-based draw put %d cards in hand, want exactly 1 (exempt)", got)
	}
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("the first draw after the exempt one drew %d cards, want 4", got)
	}
	if got := drawFor(t, g, me, 1); got != 1 {
		t.Errorf("the draw after that drew %d cards, want 1", got)
	}
}

// A second card drawn in the draw step is not the exempt one.
func TestReedRichardsSecondDrawInTheDrawStepIsReplaced(t *testing.T) {
	g, me := reedOnTable(t)
	// The fixture parks the table in seat 0's draw step with the
	// turn-based draw already completed.
	if g.Turn.Step != game.StepDraw || g.Turn.ActiveSeat != 0 {
		t.Fatalf("fixture is at step %v, seat %d", g.Turn.Step, g.Turn.ActiveSeat)
	}
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("an extra draw in the draw step drew %d cards, want 4", got)
	}
}

// On an opponent's turn the exemption does not apply (it is for your own
// draw steps): your first draw there is four.
func TestReedRichardsFirstDrawOnAnOpponentsTurnIsFour(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() {
		g.Turn.ActiveSeat = 1
		g.Turn.Step = game.StepDraw
	})
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("a draw in an opponent's draw step drew %d cards, want 4", got)
	}
}

func TestReedRichardsLeavesAnOpponentsDrawAlone(t *testing.T) {
	g, me := reedOnTable(t)
	_ = me
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	if got := drawFor(t, g, g.Seats[1], 1); got != 1 {
		t.Errorf("an opponent drew %d cards, want 1", got)
	}
}

// The mark is the turn's: next turn the first draw is four again.
func TestReedRichardsMarkEndsWithTheTurn(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Fatalf("first draw drew %d cards, want 4", got)
	}
	nextTurnOfMine(t, g)
	// The turn-based draw of the new turn was exempt; the next draw is the
	// new turn's first time.
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("the next turn's first draw drew %d cards, want 4", got)
	}
}

// Each Reed is its own effect with its own mark (CR 614.5): one that
// arrives after the first has had its time still has its own first time.
func TestReedRichardsEachReedHasItsOwnFirstTime(t *testing.T) {
	g, me := reedOnTable(t)
	g.WithWriteLock(func() { g.Turn.Step = game.StepPrecombatMain })
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Fatalf("first draw drew %d cards, want 4", got)
	}
	// A second Reed arrives after the first has had its time.
	pushCatalogPermanent(g, me.ID, "Reed Richards, Smartest Man", "Legendary Creature — Human Scientist Hero", reedRichardsOracle, false)
	if got := drawFor(t, g, me, 1); got != 4 {
		t.Errorf("a new Reed has its own first time: drew %d cards, want 4", got)
	}
	if got := drawFor(t, g, me, 1); got != 1 {
		t.Errorf("both Reeds have been used: drew %d cards, want 1", got)
	}
}
