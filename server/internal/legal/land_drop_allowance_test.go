package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/actions"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// land_drop_allowance_test.go — #2203: the enumerator offers a land
// play exactly while the seat has a land drop left (CR 305.2), with the
// allowance raised by a controlled Exploration or a one-turn grant
// (#500). The ready ring on a hand land is this list's digest, so a land
// offered here past the allowance is a land that glows when it cannot
// be played.

const oracleExploration2203 = "0c2841bb-038c-4fbf-8360-bc0a1522b58d"

// landDropTable is the active seat in its precombat main phase with
// three basic lands in hand and nothing else.
func landDropTable(t *testing.T) (*game.Game, *game.Player, []uuid.UUID) {
	t.Helper()
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	lands := []uuid.UUID{
		handCard(active, basic("Forest", "Forest")),
		handCard(active, basic("Island", "Island")),
		handCard(active, basic("Mountain", "Mountain")),
	}
	advanceTo(t, g, game.StepPrecombatMain)
	return g, active, lands
}

// playOfferedLand dispatches the enumerator's own land move for card,
// failing if none is offered.
func playOfferedLand(t *testing.T, g *game.Game, seat, card uuid.UUID) {
	t.Helper()
	moves := landMovesFor(g, seat, card)
	if len(moves) != 1 {
		t.Fatalf("want one land move for the card, got %v", labels(legal.EnumerateFor(g, seat)))
	}
	m := moves[0]
	a := actions.Action{Type: actions.Type(m.Type), Player: m.Player, Caller: seat, Params: m.Params}
	if err := actions.Dispatch(g, a); err != nil {
		t.Fatalf("play the offered land: %v", err)
	}
}

func landMoveCount(g *game.Game, seat uuid.UUID) int {
	return countKind(legal.EnumerateFor(g, seat), legal.KindLand)
}

func TestNoLandPlayIsOfferedOnceTheDropIsSpent(t *testing.T) {
	g, active, lands := landDropTable(t)
	if got := landMoveCount(g, active.ID); got != 3 {
		t.Fatalf("before any land: want 3 land moves, got %d", got)
	}
	playOfferedLand(t, g, active.ID, lands[0])
	if got := landMoveCount(g, active.ID); got != 0 {
		t.Errorf("after the one land drop: want no land move, got %v", labels(legal.EnumerateFor(g, active.ID)))
	}
}

func TestExplorationKeepsOneMoreLandPlayOffered(t *testing.T) {
	g, active, lands := landDropTable(t)
	battlefieldCard(g, active, game.Card{
		Name: "Exploration", TypeLine: "Enchantment", ManaCost: "{G}", OracleID: oracleExploration2203,
	})
	if got := g.LandDropsRemainingFor(active.ID); got != 2 {
		t.Fatalf("setup: Exploration should allow 2 land drops, got %d", got)
	}
	playOfferedLand(t, g, active.ID, lands[0])
	moves := legal.EnumerateFor(g, active.ID)
	if got := countKind(moves, legal.KindLand); got != 2 {
		t.Fatalf("after one land with Exploration: want the 2 other lands offered, got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
	playOfferedLand(t, g, active.ID, lands[1])
	if got := landMoveCount(g, active.ID); got != 0 {
		t.Errorf("after two lands with Exploration: want no land move, got %v", labels(legal.EnumerateFor(g, active.ID)))
	}
}

func TestAOneTurnGrantKeepsOneMoreLandPlayOffered(t *testing.T) {
	g, active, lands := landDropTable(t)
	// Explore's "you may play an additional land this turn".
	g.WithWriteLock(func() { g.GrantAdditionalLandPlayForEffect(active.ID, 1) })
	playOfferedLand(t, g, active.ID, lands[0])
	moves := legal.EnumerateFor(g, active.ID)
	if got := countKind(moves, legal.KindLand); got != 2 {
		t.Fatalf("after one land with a one-turn grant: want the 2 other lands offered, got %v", labels(moves))
	}
	dispatchAll(t, g, active.ID, moves)
	playOfferedLand(t, g, active.ID, lands[1])
	if got := landMoveCount(g, active.ID); got != 0 {
		t.Errorf("after two lands with a one-turn grant: want no land move, got %v", labels(legal.EnumerateFor(g, active.ID)))
	}
}
