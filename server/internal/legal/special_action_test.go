package legal_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// special_action_test.go — the enumerator half of the CR 116.2 verb
// (#658, #659, ADR 0062 Decision 4).
//
// #544's invariant, one verb over: offer exactly the special actions
// the engine accepts, and nothing the engine would refuse. The row
// that earns this file is the split-second one — the two enumerators
// beside this one return early on SplitSecondActive, and copying them
// would silently make foretell illegal under a Trickbind.

// withSpecialActions stubs the catalog hook for one test.
func withSpecialActions(t *testing.T, oracle string, actions []game.SpecialAction) {
	t.Helper()
	prev := game.CatalogSpecialActions
	game.CatalogSpecialActions = func(id string) []game.SpecialAction {
		if id != oracle {
			return nil
		}
		return actions
	}
	t.Cleanup(func() { game.CatalogSpecialActions = prev })
}

// specialActionsOf picks the special-action moves for one source.
func specialActionsOf(moves []legal.Move, source uuid.UUID) []legal.Move {
	return movesOfKindFor(moves, legal.KindSpecialAction, source)
}

func foretellCard(oracle string) game.Card {
	return game.Card{
		Name:     "Saw It Coming",
		TypeLine: "Instant",
		ManaCost: "{1}{U}{U}",
		OracleID: oracle,
	}
}

// Foretell is enumerated on its owner's turn when the {2} is payable,
// and not when it is not — the #544 rule.
func TestForetellIsEnumeratedOnlyWhenPayable(t *testing.T) {
	const oracle = "legal-foretell-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, oracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	card := handCard(active, foretellCard(oracle))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card); len(acts) != 0 {
		t.Fatalf("a seat with no mana: want no foretell move, got %v", labels(acts))
	}

	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	moves := legal.EnumerateFor(g, active.ID)
	acts := specialActionsOf(moves, card)
	if len(acts) != 1 {
		t.Fatalf("want exactly one foretell move, got %v", labels(acts))
	}
	if acts[0].Type != legal.TypeSpecialAction {
		t.Errorf("move type = %q, want %q", acts[0].Type, legal.TypeSpecialAction)
	}
	// Every enumerated move is one the dispatcher accepts.
	dispatchAll(t, g, active.ID, moves)
}

// CR 702.61b, and the row this whole file earns. Every other
// enumerator in this package opens with
// `if g.SplitSecondActive { return }`; this one must not, because a
// special action is neither a cast nor an activation and CR 702.61b
// stops only those. A Trickbind on the stack does not stop a player
// foretelling.
func TestSplitSecondDoesNotStopForetell(t *testing.T) {
	const oracle = "legal-split-second-foretell-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	clearHand(active)
	withSpecialActions(t, oracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	card := handCard(active, foretellCard(oracle))
	battlefieldCard(g, active, basic("Island", "Island"))
	battlefieldCard(g, active, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	if n := len(specialActionsOf(legal.EnumerateFor(g, active.ID), card)); n != 1 {
		t.Fatalf("with no split second: want the foretell move, got %d", n)
	}

	g.SplitSecondActive = true
	acts := specialActionsOf(legal.EnumerateFor(g, active.ID), card)
	if len(acts) != 1 {
		t.Fatalf("under split second: want foretell still offered, got %v", labels(acts))
	}
	// And the enumerator really is under split second: the ordinary
	// cast of the same card is gone.
	if n := len(movesOfKindFor(legal.EnumerateFor(g, active.ID), legal.KindCast, card)); n != 0 {
		t.Errorf("split second is not actually active: %d cast moves still offered", n)
	}
}

// A card in an OPPONENT's hand is never enumerated for this seat, and
// foretell is never offered on a turn that is not yours (CR 702.143a).
func TestForetellIsNotEnumeratedOnAnotherSeatsTurn(t *testing.T) {
	const oracle = "legal-offturn-foretell-oracle"
	g := newTable(t)
	active := g.Seats[g.Turn.ActiveSeat]
	var other *game.Player
	for _, p := range g.Seats {
		if p.ID != active.ID {
			other = p
			break
		}
	}
	clearHand(other)
	withSpecialActions(t, oracle, []game.SpecialAction{{
		Kind: game.SpecialActionForetell, Cost: "{2}", CastCost: "{1}{U}", Label: "Foretell {2}",
	}})
	card := handCard(other, foretellCard(oracle))
	battlefieldCard(g, other, basic("Island", "Island"))
	battlefieldCard(g, other, basic("Island", "Island"))
	advanceTo(t, g, game.StepPrecombatMain)

	if acts := specialActionsOf(legal.EnumerateFor(g, other.ID), card); len(acts) != 0 {
		t.Errorf("foretell offered on another seat's turn: %v", labels(acts))
	}
}
