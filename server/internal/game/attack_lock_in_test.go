package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// attack_lock_in_test.go covers #3621 and #3622 — the staged attack
// declaration is CLOSED once it has been locked in (CR 508.1), and it
// is locked in before priority can reach any other seat.

// reopenAttackDeclarationForTest starts a fresh declaration inside the
// same step, for the older tests that declare attackers in several
// separate rounds. Production reopens a declaration only by clearing
// combat.
func reopenAttackDeclarationForTest(g *Game) {
	g.WithWriteLock(func() { g.attacksDeclared = false })
}

// TestDeclareAttackerIsRefusedAfterTheLockIn — #3621. A declaration
// that has been announced (its triggers on the stack or already
// resolved) is over: a second creature declared now would be tapped
// and announced separately, with its triggers firing late.
func TestDeclareAttackerIsRefusedAfterTheLockIn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if n := len(attackDeclEvents(g, a)); n != 1 {
		t.Fatalf("setup: %d attack events, want 1", n)
	}

	err := g.DeclareAttacker(b, g.Seats[1].ID)
	if !errors.Is(err, ErrAttackersAlreadyDeclared) {
		t.Fatalf("late DeclareAttacker err = %v, want ErrAttackersAlreadyDeclared", err)
	}
	if card := findCard(g, b); card.AttackingTarget != uuid.Nil || card.Tapped {
		t.Error("a refused late declaration staged or tapped the creature")
	}
	if n := len(attackDeclEvents(g, b)); n != 0 {
		t.Errorf("a refused late declaration was announced %d times", n)
	}
}

// TestBulkDeclareAttackersIsRefusedAfterTheLockIn — the bulk verb is
// the same declaration and closes with it.
func TestBulkDeclareAttackersIsRefusedAfterTheLockIn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)

	_, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: b, Target: g.Seats[1].ID}})
	if !errors.Is(err, ErrAttackersAlreadyDeclared) {
		t.Fatalf("late DeclareAttackers err = %v, want ErrAttackersAlreadyDeclared", err)
	}
	if card := findCard(g, b); card.AttackingTarget != uuid.Nil || card.Tapped {
		t.Error("a refused late bulk declaration staged or tapped the creature")
	}
}

// TestStagedDeclarationStaysOpenUntilTheLockIn — the other half: the
// sandbox's click-by-click staging and re-point are untouched before
// the lock-in.
func TestStagedDeclarationStaysOpenUntilTheLockIn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker a: %v", err)
	}
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point a: %v", err)
	}
	if err := g.DeclareAttacker(b, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker b before the lock-in: %v", err)
	}
}

// TestUnmetRequirementKeepsTheDeclarationOpenAtLockIn — a lock-in that
// happens while a CR 508.1d requirement is still owed must not close
// the declaration, or the pass (refused until the requirement is met)
// and the declaration (refused as closed) would wedge the table.
func TestUnmetRequirementKeepsTheDeclarationOpenAtLockIn(t *testing.T) {
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	zurgo := mustAttackCreature(t, g, me, 1)
	bear := pushKeywordCreature(t, g, me, 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(bear, opp.ID); err != nil {
		t.Fatalf("declare the Bear: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if err := g.DeclareAttacker(zurgo, opp.ID); err != nil {
		t.Fatalf("the creature that must attack was refused after a lock-in with its requirement owed: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the requirement obeyed: %v", err)
	}
}

// TestActivePlayersPassLocksTheDeclarationInFirst — #3622. The active
// player's pass is the declaration being done. The attack is announced
// BEFORE priority reaches the next seat, so nothing that seat casts can
// be logged ahead of the attack it answers.
func TestActivePlayersPassLocksTheDeclarationInFirst(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, _ := attackSetup(t, g)

	if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if n := len(attackDeclEvents(g, a)); n != 0 {
		t.Fatalf("setup: announced before the pass (%d)", n)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if g.Turn.PriorityHolder != 1 {
		t.Fatalf("priority holder = %d, want the next seat (1)", g.Turn.PriorityHolder)
	}
	if n := len(attackDeclEvents(g, a)); n != 1 {
		t.Fatalf("the pass must announce the attack before priority moves: %d attack events", n)
	}
	// And the declaration is closed to the active player from here.
	b := pushPlainCreature(t, g, g.Seats[0], "Late")
	if err := g.DeclareAttacker(b, g.Seats[1].ID); !errors.Is(err, ErrAttackersAlreadyDeclared) {
		t.Fatalf("declaring after the pass: err = %v, want ErrAttackersAlreadyDeclared", err)
	}
}

// TestActivePlayersPassKeepsPriorityWhenTheLockInStacksATrigger —
// CR 508.2: when the declaration puts triggers on the stack, the
// active player receives priority with them. The pass that locked the
// declaration in is consumed by it, as the priority wrap's is.
func TestActivePlayersPassKeepsPriorityWhenTheLockInStacksATrigger(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-lock-in-oracle"
	attacker := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: attacker,
		Name:       "Test Attacker",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	withCatalogTriggers(t, func(id string) []TriggeredAbility {
		if id != oracle {
			return nil
		}
		return []TriggeredAbility{{
			Watches: []EventKind{EventAttack},
			AppliesTo: func(ev Event, source *Card, _ Characteristic, _ *Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(ev Event, source *Card, _ Characteristic, _ *Game) *StackItem {
				return NewTriggeredItem(source, "Test attack trigger")
			},
		}}
	})
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("PassPriority: %v", err)
	}
	if len(g.StackMeta) == 0 {
		t.Fatal("the attack trigger did not reach the stack")
	}
	if g.Turn.PriorityHolder != g.Turn.ActiveSeat {
		t.Fatalf("priority holder = %d, want the active seat with the trigger on the stack", g.Turn.PriorityHolder)
	}
}
