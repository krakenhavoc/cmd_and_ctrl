package actions

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// remove_from_combat_test.go covers the remove_from_combat verb (#672,
// CR 506.4) at the dispatch layer. The primitive's rules cases (what a
// removed attacker or blocker does to the other side, combat damage)
// live in game/remove_from_combat_test.go.

func removeParams(t *testing.T, id string) []byte {
	t.Helper()
	return params(t, map[string]string{"instance_id": id})
}

func attackWith(t *testing.T, g *game.Game, ids ...string) {
	t.Helper()
	advanceTo(t, g, game.StepDeclareAttackers)
	for _, id := range ids {
		act := mustAction(t, TypeDeclareAttacker, params(t, map[string]string{
			"attacker": id,
			"target":   g.Seats[1].ID.String(),
		}))
		act.Caller = g.Seats[0].ID
		if err := Dispatch(g, act); err != nil {
			t.Fatalf("declare_attacker: %v", err)
		}
	}
}

func isAttacking(g *game.Game, id string) bool {
	for _, c := range g.Battlefield.Cards {
		if c.InstanceID.String() == id {
			return c.AttackingTarget != uuid.Nil
		}
	}
	return false
}

// Removing one attacker leaves the rest of the declaration alone —
// the whole point of the verb over clear_combat.
func TestDispatchRemoveFromCombatLeavesOtherAttackers(t *testing.T) {
	g := newGame(t)
	a := pushCreature(t, g, g.Seats[0])
	b := pushCreature(t, g, g.Seats[0])
	attackWith(t, g, a, b)

	act := mustAction(t, TypeRemoveFromCombat, removeParams(t, a))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); err != nil {
		t.Fatalf("remove_from_combat: %v", err)
	}
	if isAttacking(g, a) {
		t.Errorf("the removed creature is still attacking")
	}
	if !isAttacking(g, b) {
		t.Errorf("the other attacker stopped attacking")
	}
}

// A seat cannot pull another seat's creature out of combat; an admin
// (Caller == uuid.Nil) can, to fix a wedged board.
func TestDispatchRemoveFromCombatControllerGate(t *testing.T) {
	g := newGame(t)
	a := pushCreature(t, g, g.Seats[0])
	attackWith(t, g, a)

	foreign := mustAction(t, TypeRemoveFromCombat, removeParams(t, a))
	foreign.Caller = g.Seats[1].ID
	if err := Dispatch(g, foreign); !errors.Is(err, game.ErrCardCallerMismatch) {
		t.Fatalf("foreign caller: got %v, want ErrCardCallerMismatch", err)
	}
	if !isAttacking(g, a) {
		t.Fatalf("a refused remove_from_combat still took the creature out of combat")
	}

	admin := mustAction(t, TypeRemoveFromCombat, removeParams(t, a))
	if err := Dispatch(g, admin); err != nil {
		t.Fatalf("admin remove_from_combat: %v", err)
	}
	if isAttacking(g, a) {
		t.Errorf("the admin's removal did not take effect")
	}
}

// A permanent that is on the battlefield but not in combat is a
// refusal, so the verb never mints an undo entry for a no-op.
func TestDispatchRemoveFromCombatRefusesAPermanentNotInCombat(t *testing.T) {
	g := newGame(t)
	a := pushCreature(t, g, g.Seats[0])
	advanceTo(t, g, game.StepDeclareAttackers)

	act := mustAction(t, TypeRemoveFromCombat, removeParams(t, a))
	act.Caller = g.Seats[0].ID
	if err := Dispatch(g, act); !errors.Is(err, game.ErrNotInCombat) {
		t.Fatalf("got %v, want ErrNotInCombat", err)
	}
}
