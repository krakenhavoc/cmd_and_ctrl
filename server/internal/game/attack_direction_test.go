package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// ADR 0206 (#4197): "each player may attack only the nearest opponent in
// the chosen direction and planeswalkers controlled by that opponent".
// The granted half (Teyo, Geometric Tactician's −2) is tested here; the
// derived half (Pramikon, Sky Rampart) reads the catalog and is tested
// with the card in cards/effects.

func grantDirection(g *Game, dir AttackDirection, by uuid.UUID) {
	g.WithWriteLock(func() {
		g.GrantAttackDirectionForEffect(dir, "Test Tactician", uuid.Nil, g.UntilYourNextTurnDuration(by))
	})
}

func directionErr(t *testing.T, err error) *AttackDirectionError {
	t.Helper()
	if !errors.Is(err, ErrIllegalAttackTarget) {
		t.Fatalf("err = %v, want ErrIllegalAttackTarget", err)
	}
	var de *AttackDirectionError
	if !errors.As(err, &de) {
		t.Fatalf("refusal = %T %v, want *AttackDirectionError", err, err)
	}
	return de
}

// Left is the next seat in turn order and right the previous one, and a
// seat whose player has left the game is skipped (CR 101.4, 800.4a).
func TestNearestOpponentWalksTheTableAndSkipsPlayersWhoLeft(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	s := g.Seats
	for i, p := range s {
		if got, want := g.NearestOpponentForEffect(p.ID, AttackDirectionLeft), s[(i+1)%4].ID; got != want {
			t.Errorf("seat %d left = %s, want seat %d", i, got, (i+1)%4)
		}
		if got, want := g.NearestOpponentForEffect(p.ID, AttackDirectionRight), s[(i+3)%4].ID; got != want {
			t.Errorf("seat %d right = %s, want seat %d", i, got, (i+3)%4)
		}
	}
	s[1].Eliminated = true
	if got := g.NearestOpponentForEffect(s[0].ID, AttackDirectionLeft); got != s[2].ID {
		t.Errorf("with seat 1 gone, seat 0's left = %s, want seat 2", got)
	}
	if got := g.NearestOpponentForEffect(s[2].ID, AttackDirectionRight); got != s[0].ID {
		t.Errorf("with seat 1 gone, seat 2's right = %s, want seat 0", got)
	}
	if got := g.NearestOpponentForEffect(s[0].ID, ""); got != uuid.Nil {
		t.Errorf("no direction named %s", got)
	}
}

// Under a left rule a player may attack only the next player and that
// player's planeswalkers: no other opponent, no other planeswalker and
// no battle, and the verbs agree with the enumerator's list.
func TestAttackDirectionAllowsOnlyTheNearestOpponentAndTheirPlaneswalkers(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	s := g.Seats
	nextWalker := pushPlaneswalkerForTest(g, s[2].ID, "Next Walker", 3)
	farWalker := pushPlaneswalkerForTest(g, s[3].ID, "Far Walker", 3)
	nextBattle := pushBattleForTest(g, s[3].ID, s[2].ID, "Next's Battle", 4)
	grantDirection(g, AttackDirectionLeft, s[0].ID)
	// A creature that arrives after the grant is held to it too: the
	// rule is on the player.
	bear := readyCreatureFor(g, s[1].ID, "Late Bear")
	advanceToDeclareOf(t, g, 1)

	got := attackTargetsOf(g, bear)
	if want := []uuid.UUID{s[2].ID, nextWalker}; len(got) != 2 || !slices.Contains(got, want[0]) || !slices.Contains(got, want[1]) {
		t.Fatalf("per-attacker targets = %v, want only seat 2 and its planeswalker %v", got, want)
	}
	for _, refused := range []uuid.UUID{s[3].ID, s[0].ID, farWalker, nextBattle} {
		de := directionErr(t, g.Clone().DeclareAttacker(bear, refused))
		if de.Nearest != s[2].ID || de.Direction != AttackDirectionLeft {
			t.Errorf("refusal at %s = %+v, want left naming seat 2", refused, de)
		}
	}
	de := directionErr(t, g.Clone().DeclareAttacker(bear, s[3].ID))
	want := "Late Bear may attack only " + s[2].Name + ", the nearest opponent to the left, and planeswalkers " +
		s[2].Name + " controls (Test Tactician)."
	if got := de.Sentence(); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if _, err := g.Clone().DeclareAttackers([]AttackDeclaration{{Attacker: bear, Target: s[3].ID}}); err == nil {
		t.Error("the bulk verb accepted an attack off the direction")
	}
	for _, ok := range []uuid.UUID{s[2].ID, nextWalker} {
		if err := g.Clone().DeclareAttacker(bear, ok); err != nil {
			t.Errorf("declare at %s: %v", ok, err)
		}
	}
}

// "Until your next turn": it holds through every other player's turn
// and ends as the granting player's next turn begins (CR 611.2b), read
// live.
func TestAttackDirectionGrantLastsUntilTheGrantersNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	s := g.Seats
	grantDirection(g, AttackDirectionRight, s[0].ID)
	bears := []uuid.UUID{
		readyCreatureFor(g, s[0].ID, "Bear 0"),
		readyCreatureFor(g, s[1].ID, "Bear 1"),
		readyCreatureFor(g, s[2].ID, "Bear 2"),
		readyCreatureFor(g, s[3].ID, "Bear 3"),
	}
	for seat := 1; seat < 4; seat++ {
		advanceToDeclareOf(t, g, seat)
		got := attackTargetsOf(g, bears[seat])
		if want := s[seat-1].ID; len(got) != 1 || got[0] != want {
			t.Errorf("seat %d targets = %v, want only seat %d to the right", seat, got, seat-1)
		}
	}
	advanceToDeclareOf(t, g, 0)
	if got := attackTargetsOf(g, bears[0]); len(got) != 3 {
		t.Errorf("on the granter's next turn targets = %v, want all three opponents", got)
	}
}

// Two rules that disagree leave a player at a four-player table nobody
// to attack; with two players left both directions name the same
// opponent, so they agree again.
func TestAttackDirectionsThatDisagreeForbidEveryAttack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	s := g.Seats
	grantDirection(g, AttackDirectionLeft, s[0].ID)
	grantDirection(g, AttackDirectionRight, s[0].ID)
	bear := readyCreatureFor(g, s[1].ID, "Bear")
	advanceToDeclareOf(t, g, 1)
	if got := attackTargetsOf(g, bear); len(got) != 0 {
		t.Fatalf("targets = %v, want none: left and right disagree", got)
	}
	s[3].Eliminated = true
	if got := attackTargetsOf(g, bear); len(got) != 0 {
		t.Errorf("three players left: targets = %v, want none (left is seat 2, right is seat 0)", got)
	}
	s[0].Eliminated = true
	if got := attackTargetsOf(g, bear); len(got) != 1 || got[0] != s[2].ID {
		t.Errorf("two players left: targets = %v, want seat 2 both ways", got)
	}
}

// CR 508.1d counts requirements "without disobeying any restrictions":
// a creature goaded by its controller's nearest opponent under a
// direction rule may attack only that opponent, so it owes an attack at
// the goader and nothing else.
func TestAttackDirectionNarrowsWhatAGoadedCreatureOwes(t *testing.T) {
	tb := newGoadTable(t)
	g := tb.g
	goad(t, g, tb.x, tb.a.ID)
	grantDirection(g, AttackDirectionLeft, tb.c.ID)
	advanceIntoStep(t, g, StepDeclareAttackers)

	requirementErr(t, g.Clone().PassPriority())
	directionErr(t, g.Clone().DeclareAttacker(tb.x, tb.b.ID))
	if err := g.DeclareAttacker(tb.x, tb.a.ID); err != nil {
		t.Fatalf("goaded creature at its goader, the only player it may attack: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the only possible attack declared: %v", err)
	}
}

// The grant is plain data on PlayerStatic and survives a snapshot.
func TestAttackDirectionGrantSurvivesASnapshot(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	s := g.Seats
	grantDirection(g, AttackDirectionLeft, s[0].ID)
	_, restored := roundTrip(t, g)
	var got []AttackDirection
	for _, st := range restored.Seats[2].Statics {
		got = append(got, st.AttackDirection)
	}
	if len(got) != 1 || got[0] != AttackDirectionLeft {
		t.Errorf("restored seat 2 statics directions = %v, want [left]", got)
	}
}
