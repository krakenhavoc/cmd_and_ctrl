package game

import (
	"errors"
	"slices"
	"testing"

	"github.com/google/uuid"
)

// #2109 (ADR 0063's amendment of 2026-10-08): a player who can't attack
// another player or their permanents during their next turn.

func readyCreatureFor(g *Game, controller uuid.UUID, name string) uuid.UUID {
	id := pushTypedTestCard(g, Card{
		Name: name, TypeLine: "Creature — Test", Power: 2, Toughness: 2,
		Owner: controller, Controller: controller,
	})
	g.WithWriteLock(func() { findBattlefieldCard(g, id).SummonedThisTurn = false })
	return id
}

// advanceToDeclareOf walks the game until it is `seat`'s declare
// attackers step, crossing turns as needed.
func advanceToDeclareOf(t *testing.T, g *Game, seat int) {
	t.Helper()
	for steps := 0; g.Turn.ActiveSeat != seat || g.Turn.Step != StepDeclareAttackers; steps++ {
		if steps > 400 {
			t.Fatalf("never reached seat %d's declare attackers", seat)
		}
		if _, err := g.AdvanceStep(); err != nil {
			t.Fatalf("AdvanceStep: %v", err)
		}
	}
}

func grantCantAttack(g *Game, attacker, protected uuid.UUID) {
	g.WithWriteLock(func() {
		g.GrantCantAttackPlayerForEffect(attacker, protected, "Test grant", uuid.Nil)
	})
}

func TestCantAttackPlayerNextTurnRefusesThePlayerAndTheirPermanents(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	protected, restricted, third := g.Seats[0], g.Seats[1], g.Seats[2]
	walker := pushPlaneswalkerForTest(g, protected.ID, "Protected Walker", 3)
	otherWalker := pushPlaneswalkerForTest(g, third.ID, "Third Walker", 3)
	// The grant is made on the protected player's turn (seat 0 is active).
	grantCantAttack(g, restricted.ID, protected.ID)
	// A creature that arrives after the grant is held back as well: the
	// restriction is on the player, not on the creatures they had.
	late := readyCreatureFor(g, restricted.ID, "Late Arrival")
	advanceToDeclareOf(t, g, 1)

	got := attackTargetsOf(g, late)
	for _, refused := range []uuid.UUID{protected.ID, walker} {
		if slices.Contains(got, refused) {
			t.Errorf("per-attacker targets %v include the forbidden %s", got, refused)
		}
	}
	for _, want := range []uuid.UUID{third.ID, g.Seats[3].ID, otherWalker} {
		if !slices.Contains(got, want) {
			t.Errorf("per-attacker targets %v miss %s", got, want)
		}
	}

	for _, target := range []uuid.UUID{protected.ID, walker} {
		err := g.Clone().DeclareAttacker(late, target)
		if !errors.Is(err, ErrIllegalAttackTarget) {
			t.Fatalf("declare at %s: err = %v, want ErrIllegalAttackTarget", target, err)
		}
		var pe *PlayerCantAttackError
		if !errors.As(err, &pe) || pe.Protected != protected.ID {
			t.Fatalf("refusal = %T %v, want *PlayerCantAttackError naming the protected player", err, err)
		}
	}
	if _, err := g.Clone().DeclareAttackers([]AttackDeclaration{{Attacker: late, Target: protected.ID}}); !errors.Is(err, ErrNoLegalAttackers) {
		t.Errorf("bulk declaration at the protected player: err = %v, want ErrNoLegalAttackers", err)
	}
	for _, ok := range []uuid.UUID{third.ID, otherWalker} {
		if err := g.Clone().DeclareAttacker(late, ok); err != nil {
			t.Errorf("declare at %s: %v", ok, err)
		}
	}
}

// The window ends with the restricted player's next turn, and it is
// read live (no sweep is needed to end it).
func TestCantAttackPlayerLastsOnlyThroughTheirNextTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	protected, restricted := g.Seats[0], g.Seats[1]
	grantCantAttack(g, restricted.ID, protected.ID)
	c := readyCreatureFor(g, restricted.ID, "Raider")

	advanceToDeclareOf(t, g, 1)
	if slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Fatal("restricted during their next turn: the protected player is offered")
	}
	// A full round later it is their turn again, and it is over.
	advanceToDeclareOf(t, g, 2)
	advanceToDeclareOf(t, g, 1)
	g.WithWriteLock(func() { findBattlefieldCard(g, c).SummonedThisTurn = false })
	if !slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Error("the turn after their next turn the restriction should be over")
	}
	if len(g.Seats[1].Statics) != 0 {
		t.Errorf("expired grant not swept: %+v", g.Seats[1].Statics)
	}
}

// A grant made in the restricted player's OWN turn covers their NEXT
// turn, not the one in progress.
func TestCantAttackPlayerMadeInTheirOwnTurnWaitsForTheNext(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, other := g.Seats[0], g.Seats[1]
	c := readyCreatureFor(g, me.ID, "Raider")
	advanceToDeclareOf(t, g, 0)
	grantCantAttack(g, me.ID, other.ID)
	if !slices.Contains(attackTargetsOf(g, c), other.ID) {
		t.Fatal("the turn in progress must not be restricted")
	}
	advanceToDeclareOf(t, g, 1)
	advanceToDeclareOf(t, g, 0)
	g.WithWriteLock(func() { findBattlefieldCard(g, c).SummonedThisTurn = false })
	if slices.Contains(attackTargetsOf(g, c), other.ID) {
		t.Error("their next turn must be restricted")
	}
}

// The restriction names a player; a player never restricts themselves,
// and an unrelated third player is unaffected.
func TestCantAttackPlayerIgnoresSelfAndOthers(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	a, b := g.Seats[0], g.Seats[1]
	grantCantAttack(g, a.ID, a.ID)
	if len(a.Statics) != 0 {
		t.Errorf("a self-grant was stored: %+v", a.Statics)
	}
	grantCantAttack(g, a.ID, b.ID)
	c := readyCreatureFor(g, g.Seats[2].ID, "Bystander")
	advanceToDeclareOf(t, g, 2)
	if !slices.Contains(attackTargetsOf(g, c), b.ID) || !slices.Contains(attackTargetsOf(g, c), a.ID) {
		t.Error("a player with no grant should be able to attack both")
	}
}

// The reader tests the duration itself: bump the turn counter past the
// window WITHOUT running a sweep and the grant is already over.
func TestCantAttackPlayerIsDurationCheckedByTheReader(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	protected, restricted := g.Seats[0], g.Seats[1]
	grantCantAttack(g, restricted.ID, protected.ID)
	c := readyCreatureFor(g, restricted.ID, "Raider")
	g.WithWriteLock(func() { restricted.TurnsBegun++ }) // their next turn has begun
	if slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Fatal("during their next turn: the protected player is offered")
	}
	g.WithWriteLock(func() { restricted.TurnsBegun++ }) // and the one after
	if !slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Error("after the window, with no sweep run: still restricted")
	}
}

// A clone (undo) owns its own statics.
func TestCantAttackGrantCloneIsolates(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	grantCantAttack(g, g.Seats[1].ID, g.Seats[0].ID)
	snap := g.Clone()
	grantCantAttack(g, g.Seats[2].ID, g.Seats[0].ID)
	if got := len(snap.Seats[2].Statics); got != 0 {
		t.Errorf("clone saw a later grant: %d statics", got)
	}
	if got := len(snap.Seats[1].Statics); got != 1 {
		t.Errorf("clone lost the earlier grant: %d statics", got)
	}
}

// --- #2719: "this turn", and narrower scopes ---------------------------

// A this-turn grant made during the restricted player's turn applies at
// once and ends with the turn.
func TestCantAttackPlayerThisTurnAppliesNowAndEndsWithTheTurn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	protected, restricted := g.Seats[0], g.Seats[1]
	c := readyCreatureFor(g, restricted.ID, "Raider")
	advanceToDeclareOf(t, g, 1)
	g.WithWriteLock(func() {
		g.GrantCantAttackPlayerThisTurnForEffect(restricted.ID, protected.ID, CantAttackScope{}, "Test grant", uuid.Nil)
	})
	if slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Fatal("the turn in progress must be restricted")
	}
	advanceToDeclareOf(t, g, 2)
	advanceToDeclareOf(t, g, 1)
	g.WithWriteLock(func() { findBattlefieldCard(g, c).SummonedThisTurn = false })
	if !slices.Contains(attackTargetsOf(g, c), protected.ID) {
		t.Error("their next turn must not be restricted")
	}
}

// Each scope covers exactly what its clause names.
func TestCantAttackScopeCoversWhatItNames(t *testing.T) {
	type covered struct{ player, jace, walker, battle bool }
	for _, tc := range []struct {
		name  string
		scope CantAttackScope
		want  covered
	}{
		{"you or permanents you control", CantAttackScope{}, covered{true, true, true, true}},
		{"you or planeswalkers you control", CantAttackScope{PlaneswalkersOnly: true}, covered{true, true, true, false}},
		{"you", CantAttackScope{PlayerOnly: true}, covered{true, false, false, false}},
		{"Jaces you control", CantAttackScope{PlayerExempt: true, PlaneswalkersOnly: true, Subtype: "Jace"}, covered{false, true, false, false}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			g := newFourPlayerActiveGame(t)
			protected, restricted, third := g.Seats[0], g.Seats[1], g.Seats[2]
			jace := pushTypedTestCard(g, Card{
				Name: "A Jace", TypeLine: "Legendary Planeswalker — Jace",
				Owner: protected.ID, Controller: protected.ID, Counters: map[string]int{CounterLoyalty: 3},
			})
			walker := pushPlaneswalkerForTest(g, protected.ID, "Other Walker", 3)
			battle := pushTypedTestCard(g, Card{
				Name: "A Siege", TypeLine: "Battle — Siege", ProtectorPlayerID: third.ID,
				Owner: protected.ID, Controller: protected.ID, Counters: map[string]int{CounterDefense: 3},
			})
			c := readyCreatureFor(g, restricted.ID, "Raider")
			advanceToDeclareOf(t, g, 1)
			g.WithWriteLock(func() {
				g.GrantCantAttackPlayerThisTurnForEffect(restricted.ID, protected.ID, tc.scope, "Test grant", uuid.Nil)
			})
			got := attackTargetsOf(g, c)
			for target, want := range map[uuid.UUID]bool{protected.ID: tc.want.player, jace: tc.want.jace, walker: tc.want.walker, battle: tc.want.battle} {
				if slices.Contains(got, target) == want {
					t.Errorf("target %s: covered = %v, want %v (offered %v)", target, !want, want, got)
				}
			}
			if !slices.Contains(got, third.ID) {
				t.Error("a third player is never covered")
			}
		})
	}
}

// The refusal says what the scope covers.
func TestPlayerCantAttackErrorSentenceFollowsTheScope(t *testing.T) {
	for scope, want := range map[CantAttackScope]string{
		{}:                        "Raider can't attack Alice or permanents they control this turn (Src).",
		{PlaneswalkersOnly: true}: "Raider can't attack Alice or planeswalkers they control this turn (Src).",
		{PlayerOnly: true}:        "Raider can't attack Alice this turn (Src).",
		{PlayerExempt: true, PlaneswalkersOnly: true, Subtype: "Jace"}: "Raider can't attack Jaces Alice controls this turn (Src).",
	} {
		e := &PlayerCantAttackError{AttackerName: "Raider", ProtectedName: "Alice", SourceName: "Src", Scope: scope}
		if got := e.Sentence(); got != want {
			t.Errorf("%+v: %q, want %q", scope, got, want)
		}
	}
}

// An attack on a planeswalker records who controlled it then, and the
// "attacked you or a planeswalker you control" read uses that record
// after the planeswalker is gone. An attack on a battle is neither.
func TestAttackedYouOrYourPlaneswalkerThisTurnReadsTheDeclaration(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	you, attacker, third := g.Seats[0], g.Seats[1], g.Seats[2]
	walker := pushPlaneswalkerForTest(g, you.ID, "Walker", 3)
	battle := pushTypedTestCard(g, Card{
		Name: "A Siege", TypeLine: "Battle — Siege", ProtectorPlayerID: third.ID,
		Owner: you.ID, Controller: you.ID, Counters: map[string]int{CounterDefense: 3},
	})
	c := readyCreatureFor(g, attacker.ID, "Raider")
	advanceToDeclareOf(t, g, 1)
	attacked := func() (b bool) {
		g.WithWriteLock(func() { b = g.AttackedYouOrYourPlaneswalkerThisTurn(attacker.ID, you.ID) })
		return
	}
	if attacked() {
		t.Fatal("nothing has attacked yet")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventAttack, Actor: attacker.ID, CardID: c, Target: third.ID})
		g.EmitEvent(Event{Kind: EventAttack, Actor: attacker.ID, CardID: c, Target: battle})
	})
	if attacked() {
		t.Fatal("an attack on another player or on a battle you control is not an attack on you")
	}
	g.WithWriteLock(func() {
		g.EmitEvent(Event{Kind: EventAttack, Actor: attacker.ID, CardID: c, Target: walker})
		g.Battlefield.Remove(walker)
	})
	if !attacked() {
		t.Error("an attack on your planeswalker counts after it has left")
	}
	if got := g.TurnTally.Attacks[len(g.TurnTally.Attacks)-1].PlaneswalkerController; got != you.ID {
		t.Errorf("record's planeswalker controller = %s, want %s", got, you.ID)
	}
	g.WithWriteLock(func() {
		if g.AttackedYouOrYourPlaneswalkerThisTurn(third.ID, you.ID) {
			t.Error("a player who is not the active player attacked nobody this turn")
		}
	})
}
