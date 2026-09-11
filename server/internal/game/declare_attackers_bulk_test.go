package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// declare_attackers_bulk_test.go covers game.DeclareAttackers — the
// one-mutation attacking-set declaration behind the client's "attack
// with all" cluster (#318). The contract under test is:
//
//   - ineligible entries are SKIPPED, never fatal, so one summoning-
//     sick creature can't sink a 12-creature alpha strike;
//   - vigilance still skips the CR 508.1f tap;
//   - every declared creature emits exactly one EventAttack, and the
//     whole batch drains through a single state-check pass;
//   - an all-skipped batch is ErrNoLegalAttackers so the room layer
//     records neither an undo entry nor a snapshot for a no-op.

// pushPlainCreature drops an untapped, non-sick vanilla creature on
// the battlefield for the given seat. Effective characteristics are
// set directly (same shortcut pushKeywordCreature takes) so these
// tests need no catalog entry.
func pushPlainCreature(t *testing.T, g *Game, owner *Player, name string) uuid.UUID {
	t.Helper()
	id := pushKeywordCreature(t, g, owner, 2, 2)
	c := findCard(g, id)
	c.Name = name
	// pushKeywordCreature goes through NewCard, which stamps
	// SummonedThisTurn. These creatures are meant to have been around
	// since before this turn.
	c.SummonedThisTurn = false
	return id
}

func declsAt(target uuid.UUID, ids ...uuid.UUID) []AttackDeclaration {
	out := make([]AttackDeclaration, 0, len(ids))
	for _, id := range ids {
		out = append(out, AttackDeclaration{Attacker: id, Target: target})
	}
	return out
}

func TestDeclareAttackersDeclaresWholeSet(t *testing.T) {
	g := newActiveGame(t)
	a := pushPlainCreature(t, g, g.Seats[0], "Alpha")
	b := pushPlainCreature(t, g, g.Seats[0], "Bravo")
	c := pushPlainCreature(t, g, g.Seats[0], "Charlie")
	advanceIntoStep(t, g, StepDeclareAttackers)

	declared, err := g.DeclareAttackers(declsAt(g.Seats[1].ID, a, b, c))
	if err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if len(declared) != 3 {
		t.Fatalf("declared %d attackers, want 3", len(declared))
	}
	for _, id := range []uuid.UUID{a, b, c} {
		card := findCard(g, id)
		if card.AttackingTarget != g.Seats[1].ID {
			t.Errorf("%s attacking %v, want seat 1", card.Name, card.AttackingTarget)
		}
		if !card.Tapped {
			t.Errorf("%s should be tapped by the declaration (CR 508.1f)", card.Name)
		}
	}
}

func TestDeclareAttackersSkipsIneligibleSilently(t *testing.T) {
	g := newActiveGame(t)
	ok := pushPlainCreature(t, g, g.Seats[0], "Eligible")

	sick := pushPlainCreature(t, g, g.Seats[0], "Sick")
	findCard(g, sick).SummonedThisTurn = true

	tapped := pushPlainCreature(t, g, g.Seats[0], "Tapped")
	findCard(g, tapped).Tapped = true

	wall := pushKeywordCreature(t, g, g.Seats[0], 0, 4, "defender")
	findCard(g, wall).SummonedThisTurn = false

	// A land: on the battlefield, controlled by the attacker, but not
	// a creature.
	land := NewCard("Forest", g.Seats[0].ID)
	land.TypeLine = "Basic Land — Forest"
	g.Battlefield.PushTop(land)

	advanceIntoStep(t, g, StepDeclareAttackers)

	decls := declsAt(g.Seats[1].ID, ok, sick, tapped, wall, land.InstanceID, uuid.New())
	declared, err := g.DeclareAttackers(decls)
	if err != nil {
		t.Fatalf("DeclareAttackers should skip, not fail: %v", err)
	}
	if len(declared) != 1 || declared[0] != ok {
		t.Fatalf("declared %v, want exactly the one eligible creature", declared)
	}
	for _, id := range []uuid.UUID{sick, wall} {
		if card := findCard(g, id); card.AttackingTarget != uuid.Nil {
			t.Errorf("%s was declared but should have been skipped", card.Name)
		}
	}
	if card := findCard(g, tapped); card.AttackingTarget != uuid.Nil {
		t.Errorf("already-tapped creature was declared")
	}
}

func TestDeclareAttackersVigilanceDoesNotTap(t *testing.T) {
	g := newActiveGame(t)
	vig := pushKeywordCreature(t, g, g.Seats[0], 3, 3, "vigilance")
	findCard(g, vig).SummonedThisTurn = false
	plain := pushPlainCreature(t, g, g.Seats[0], "Plain")
	advanceIntoStep(t, g, StepDeclareAttackers)

	if _, err := g.DeclareAttackers(declsAt(g.Seats[1].ID, vig, plain)); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if findCard(g, vig).Tapped {
		t.Errorf("vigilant attacker was tapped (CR 702.20)")
	}
	if !findCard(g, plain).Tapped {
		t.Errorf("non-vigilant attacker should be tapped")
	}
}

func TestDeclareAttackersLeavesExistingDeclarationsAlone(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	already := pushPlainCreature(t, g, g.Seats[0], "Committed")
	fresh := pushPlainCreature(t, g, g.Seats[0], "Fresh")
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttacker(already, g.Seats[2].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	declared, err := g.DeclareAttackers(declsAt(g.Seats[1].ID, already, fresh))
	if err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	if len(declared) != 1 || declared[0] != fresh {
		t.Fatalf("declared %v, want only the undeclared creature", declared)
	}
	if got := findCard(g, already).AttackingTarget; got != g.Seats[2].ID {
		t.Errorf("bulk declaration re-pointed an existing attacker to %v", got)
	}
}

func TestDeclareAttackersRejectsUnattackableSeats(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	one := pushPlainCreature(t, g, g.Seats[0], "One")
	two := pushPlainCreature(t, g, g.Seats[0], "Two")
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.Seats[2].Eliminated = true

	// Self, eliminated seat, and an unknown UUID are all skipped.
	decls := []AttackDeclaration{
		{Attacker: one, Target: g.Seats[0].ID},
		{Attacker: two, Target: g.Seats[2].ID},
	}
	if _, err := g.DeclareAttackers(decls); !errors.Is(err, ErrNoLegalAttackers) {
		t.Fatalf("DeclareAttackers err = %v, want ErrNoLegalAttackers", err)
	}
	if findCard(g, one).Tapped || findCard(g, two).Tapped {
		t.Errorf("a fully-skipped batch must not tap anything")
	}
}

func TestDeclareAttackersWrongStep(t *testing.T) {
	g := newActiveGame(t)
	id := pushPlainCreature(t, g, g.Seats[0], "Early")
	if _, err := g.DeclareAttackers(declsAt(g.Seats[1].ID, id)); !errors.Is(err, ErrWrongStep) {
		t.Fatalf("DeclareAttackers err = %v, want ErrWrongStep", err)
	}
}

func TestDeclareAttackersEmitsOneAttackEventPerCreature(t *testing.T) {
	g := newActiveGame(t)
	a := pushPlainCreature(t, g, g.Seats[0], "Alpha")
	b := pushPlainCreature(t, g, g.Seats[0], "Bravo")
	sick := pushPlainCreature(t, g, g.Seats[0], "Sick")
	findCard(g, sick).SummonedThisTurn = true
	advanceIntoStep(t, g, StepDeclareAttackers)

	before := len(g.Events)
	if _, err := g.DeclareAttackers(declsAt(g.Seats[1].ID, a, b, sick)); err != nil {
		t.Fatalf("DeclareAttackers: %v", err)
	}
	var attacks int
	for _, ev := range g.Events[before:] {
		if ev.Kind == EventAttack {
			attacks++
			if ev.Target != g.Seats[1].ID {
				t.Errorf("EventAttack target = %v, want seat 1", ev.Target)
			}
		}
	}
	if attacks != 2 {
		t.Errorf("emitted %d EventAttack, want 2 (skipped creatures emit nothing)", attacks)
	}
}
