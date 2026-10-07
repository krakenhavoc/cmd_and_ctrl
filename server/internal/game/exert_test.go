package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// exert_test.go — ADR 0130 PR 1, the engine half: exert as it attacks
// (CR 701.43). The cards are pinned in cards/effects (exert_test.go)
// and the twin moves in internal/legal.

const exertTestOracle = "exert-oracle-test"

// stubExert installs an exert-as-it-attacks ability for exertTestOracle
// and restores the hook afterwards.
func stubExert(t *testing.T, unless func(*Game, *Card) bool) {
	t.Helper()
	prev := CatalogExertOnAttack
	t.Cleanup(func() { CatalogExertOnAttack = prev })
	CatalogExertOnAttack = func(key string) *ExertOnAttack {
		if key == exertTestOracle {
			return &ExertOnAttack{Unless: unless}
		}
		return nil
	}
}

// pushExerter puts a 2/2 that may be exerted as it attacks onto the
// battlefield under `owner`, ready to attack this turn.
func pushExerter(t *testing.T, g *Game, owner *Player, keywords ...string) uuid.UUID {
	t.Helper()
	id := pushKeywordCreature(t, g, owner, 2, 2, keywords...)
	c := findCard(g, id)
	c.Name = "Exerter"
	c.OracleID = exertTestOracle
	// Printed too, so a layer recompute (a control change) keeps them.
	c.Keywords = append([]string(nil), keywords...)
	c.SummonedThisTurn = false
	return id
}

func exertEvents(g *Game, card uuid.UUID) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventExert && ev.CardID == card {
			out = append(out, ev)
		}
	}
	return out
}

func untapSkipsFor(g *Game, id uuid.UUID) []UntapSkip {
	if c := findCard(g, id); c != nil {
		return c.NextUntapSkips
	}
	return nil
}

// passTurnTo passes turns until `seat` is the active player.
func passTurnTo(t *testing.T, g *Game, seat int) {
	t.Helper()
	for i := 0; i < 2*len(g.Seats)+2; i++ {
		passTurn(t, g)
		if g.Turn.ActiveSeat == seat {
			return
		}
	}
	t.Fatalf("seat %d never became active", seat)
}

// TestExertAsItAttacksIsPaidAtTheLockIn is ADR 0130 test 1: the verb
// stages the choice and pays nothing; the lock-in records one marker
// keyed to the exerter, one TurnTally.Exerts entry and one EventExert,
// in the same event batch as the EventAttack, ahead of it.
func TestExertAsItAttacksIsPaidAtTheLockIn(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushExerter(t, g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: opp.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare with exert: %v", err)
	}
	if !findCard(g, a).ExertOnAttack {
		t.Fatal("the choice is staged on the card")
	}
	if len(exertEvents(g, a)) != 0 || len(untapSkipsFor(g, a)) != 0 || len(g.TurnTally.Exerts) != 0 {
		t.Fatal("nothing is paid before the lock-in")
	}
	lockInAttackDeclaration(t, g)

	if findCard(g, a).ExertOnAttack {
		t.Error("the staged choice is spent at the lock-in")
	}
	skips := untapSkipsFor(g, a)
	if len(skips) != 1 || skips[0].Player != me.ID || skips[0].While != nil {
		t.Fatalf("NextUntapSkips = %+v, want one marker keyed to the exerter", skips)
	}
	if len(g.TurnTally.Exerts) != 1 {
		t.Fatalf("TurnTally.Exerts = %+v, want one record", g.TurnTally.Exerts)
	}
	rec := g.TurnTally.Exerts[0]
	if rec.Object != a || rec.Player != me.ID || rec.Epoch != findCard(g, a).ObjectEpoch || rec.PhaseID != g.Turn.PhaseID {
		t.Errorf("record = %+v", rec)
	}
	exerts := exertEvents(g, a)
	attacks := attackDeclEvents(g, a)
	if len(exerts) != 1 || len(attacks) != 1 {
		t.Fatalf("%d exert events and %d attack events, want one each", len(exerts), len(attacks))
	}
	ex, at := exerts[0], attacks[0]
	if ex.Actor != me.ID || ex.Source != a || ex.Target != opp.ID {
		t.Errorf("exert event = %+v", ex)
	}
	if ex.Batch != at.Batch {
		t.Errorf("exert batch %d, attack batch %d: the exert is paid in the declaration's batch (CR 508.1j, 508.1m)", ex.Batch, at.Batch)
	}
	if ex.Seq >= at.Seq {
		t.Error("the exert is paid before the creature becomes an attacking creature (CR 508.1j before 508.1k)")
	}
	g.WithWriteLock(func() {
		if !g.ExertedThisTurn(a) {
			t.Error("ExertedThisTurn reads the record")
		}
	})
}

// TestExertedCreatureSkipsOnlyTheExertersNextUntap is ADR 0130 test 2.
func TestExertedCreatureSkipsOnlyTheExertersNextUntap(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	opp := g.Seats[1]
	a := pushExerter(t, g, g.Seats[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	if _, err := g.DeclareAttackersWith([]AttackDeclaration{{Attacker: a, Target: opp.ID, Exert: true}}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare: %v", err)
	}

	passTurnTo(t, g, 0)
	if !findCard(g, a).Tapped {
		t.Fatal("CR 701.43a: it doesn't untap during the exerter's next untap step")
	}
	if len(untapSkipsFor(g, a)) != 0 {
		t.Error("the marker is used up at that step")
	}
	passTurnTo(t, g, 0)
	if findCard(g, a).Tapped {
		t.Fatal("it untaps at the exerter's untap step after that")
	}
}

// TestExertingTwiceIsOneSkippedUntap is ADR 0130 test 3, the #2029
// guard: a second exert before the exerter's next untap step leaves ONE
// marker, and the creature untaps one step later, not two
// (CR 701.43b). Whoever makes the plain marker count must keep this
// green.
func TestExertingTwiceIsOneSkippedUntap(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushExerter(t, g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: opp.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("first exert: %v", err)
	}
	lockInAttackDeclaration(t, g)
	// A second exert before the untap step: the engine's one primitive,
	// as a cost would call it.
	g.WithWriteLock(func() {
		if !g.exertLocked(a, me.ID, uuid.Nil) {
			t.Fatal("second exert refused")
		}
	})
	if n := len(g.TurnTally.Exerts); n != 2 {
		t.Fatalf("each exert is recorded: %d records", n)
	}
	if n := len(untapSkipsFor(g, a)); n != 1 {
		t.Fatalf("CR 701.43b: two exerts are one marker, got %d", n)
	}

	passTurnTo(t, g, 0)
	if !findCard(g, a).Tapped {
		t.Fatal("it doesn't untap during the exerter's next untap step")
	}
	passTurnTo(t, g, 0)
	if findCard(g, a).Tapped {
		t.Fatal("CR 701.43b: every exert expired at the same untap step, so it untaps at the next one")
	}
}

// TestVigilantExertedCreatureStaysUntapped is ADR 0130 test 4: vigilance
// keeps it untapped (CR 702.20b), the marker is still recorded and
// expires at the exerter's untap step having done nothing.
func TestVigilantExertedCreatureStaysUntapped(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	opp := g.Seats[1]
	a := pushExerter(t, g, g.Seats[0], "vigilance")
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: opp.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if findCard(g, a).Tapped {
		t.Fatal("vigilance: attacking doesn't tap it")
	}
	if len(untapSkipsFor(g, a)) != 1 {
		t.Fatal("the exert still records its marker")
	}
	passTurnTo(t, g, 0)
	if findCard(g, a).Tapped {
		t.Error("it stays untapped")
	}
	if len(untapSkipsFor(g, a)) != 0 {
		t.Error("the marker expires at the exerter's untap step having done nothing")
	}
}

// TestBorrowedExertedCreatureUntapsForItsOwner is ADR 0130 test 5: the
// skip names the exerter's untap step (CR 701.43a), so a creature
// borrowed until end of turn and exerted untaps during its owner's.
func TestBorrowedExertedCreatureUntapsForItsOwner(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	me, owner := g.Seats[0], g.Seats[1]
	a := pushExerter(t, g, owner, "haste")
	advanceIntoStep(t, g, StepPrecombatMain)
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), a, me.ID, g.UntilEndOfTurnDuration(), "test — borrow") {
			t.Fatal("GainControlForEffect")
		}
	})
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: owner.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if s := untapSkipsFor(g, a); len(s) != 1 || s[0].Player != me.ID {
		t.Fatalf("marker = %+v, want one keyed to the borrower", s)
	}

	passTurnTo(t, g, 1)
	c := findCard(g, a)
	if c.Controller != owner.ID {
		t.Fatalf("control returned at end of turn: controller %s", c.Controller)
	}
	if c.Tapped {
		t.Fatal("it untaps during its owner's untap step (the Amonkhet ruling)")
	}
}

// TestClearedOrUndoneExertRecordsNothing is ADR 0130 test 6: a staged
// exert taken back before the lock-in — clear_combat, or undo (the
// clone taken before the verb) — exerted nothing.
func TestClearedOrUndoneExertRecordsNothing(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	opp := g.Seats[1]
	a := pushExerter(t, g, g.Seats[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	before := g.Clone()

	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: opp.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare: %v", err)
	}
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("ClearCombat: %v", err)
	}
	if findCard(g, a).ExertOnAttack {
		t.Fatal("clearing combat takes the staged choice back")
	}
	lockInAttackDeclaration(t, g)
	if len(exertEvents(g, a)) != 0 || len(untapSkipsFor(g, a)) != 0 || len(g.TurnTally.Exerts) != 0 {
		t.Fatal("a cleared staged exert recorded something")
	}

	// Undo restores the clone taken before the verb.
	if c := findCard(before, a); c.ExertOnAttack || len(c.NextUntapSkips) != 0 {
		t.Fatal("the pre-verb clone carries no exert")
	}
	lockInAttackDeclaration(t, before)
	if len(exertEvents(before, a)) != 0 || len(before.TurnTally.Exerts) != 0 {
		t.Fatal("an undone exert recorded something")
	}
}

// TestExertOnACreatureWithoutTheAbilityIsRefused is ADR 0130 test 7:
// both verbs refuse `exert: true` on a creature that can't be exerted
// as it attacks, the bulk verb for the whole batch, and nothing is
// staged or paid.
func TestExertOnACreatureWithoutTheAbilityIsRefused(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	plain := pushPlainCreature(t, g, me, "Plain")
	exerter := pushExerter(t, g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)

	err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: plain, Target: opp.ID, Exert: true}, DeclareAttackersParams{})
	if !errors.Is(err, ErrCantExert) {
		t.Fatalf("single verb: err = %v, want ErrCantExert", err)
	}
	if c := findCard(g, plain); c.AttackingTarget != uuid.Nil || c.Tapped {
		t.Fatal("a refused declaration stages nothing")
	}

	_, err = g.DeclareAttackersWith([]AttackDeclaration{
		{Attacker: exerter, Target: opp.ID, Exert: true},
		{Attacker: plain, Target: opp.ID, Exert: true},
	}, DeclareAttackersParams{})
	if !errors.Is(err, ErrCantExert) {
		t.Fatalf("bulk verb: err = %v, want ErrCantExert", err)
	}
	for _, id := range []uuid.UUID{plain, exerter} {
		if c := findCard(g, id); c.AttackingTarget != uuid.Nil || c.ExertOnAttack {
			t.Fatal("the whole batch is refused")
		}
	}

	// The Unless condition is the other refusal: an Exerter that "has
	// been exerted this turn" can't be (Combat Celebrant's shape).
	stubExert(t, func(g *Game, c *Card) bool { return g.ExertedThisTurn(c.InstanceID) })
	if _, err := g.DeclareAttackersWith([]AttackDeclaration{{Attacker: exerter, Target: opp.ID, Exert: true}}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("first exert this turn: %v", err)
	}
	lockInAttackDeclaration(t, g)
	g.WithWriteLock(func() {
		if g.canExertAsItAttacksLocked(findCard(g, exerter)) {
			t.Error("an announced attacker can't be exerted later in combat, and this one has been exerted this turn")
		}
	})
	// And re-pointing an announced attacker can't exert it (the ruling:
	// you exert as you declare it, never later in combat).
	stubExert(t, nil)
	err = g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: exerter, Target: opp.ID, Exert: true}, DeclareAttackersParams{})
	if !errors.Is(err, ErrCantExert) {
		t.Fatalf("re-point of an announced attacker with exert: err = %v, want ErrCantExert", err)
	}
}

// TestRequirementToAttackIsMetWithoutExerting is ADR 0130 test 8: a
// creature that must attack may attack without paying the optional
// cost (CR 508.1d), and paying it does not change what is owed.
func TestRequirementToAttackIsMetWithoutExerting(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	opp := g.Seats[1]
	a := pushExerter(t, g, g.Seats[0])
	registerScopedEffectForTest(t, g, a, []Mod{AddAttackRequirementMod(uuid.Nil)}, IndefiniteDuration())
	advanceIntoStep(t, g, StepDeclareAttackers)

	requirementErr(t, g.PassPriority())
	if err := g.DeclareAttacker(a, opp.ID); err != nil {
		t.Fatalf("attacking without exerting obeys the requirement: %v", err)
	}
	if err := g.PassPriority(); err != nil {
		t.Fatalf("pass with the requirement obeyed: %v", err)
	}
	if len(exertEvents(g, a)) != 0 {
		t.Error("nothing exerted it")
	}
}

// TestCreaturePutOntoTheBattlefieldAttackingCantBeExerted is ADR 0130
// test 9: it was never declared (CR 508.4, 508.3a), so it never passes
// the declaration and can't be exerted.
func TestCreaturePutOntoTheBattlefieldAttackingCantBeExerted(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushExerter(t, g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)
	g.WithWriteLock(func() { g.stampEntryAttackerLocked(a, opp.ID) })

	g.WithWriteLock(func() {
		if g.CanExertAsItAttacksForEffect(findCard(g, a)) {
			t.Error("a creature put onto the battlefield attacking is offered no exert")
		}
	})
	err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: opp.ID, Exert: true}, DeclareAttackersParams{})
	if !errors.Is(err, ErrCantExert) {
		t.Fatalf("err = %v, want ErrCantExert", err)
	}
	lockInAttackDeclaration(t, g)
	if len(exertEvents(g, a)) != 0 {
		t.Error("it was exerted")
	}
}

// TestExertDoesNothingOffTheBattlefield — CR 701.43c.
func TestExertDoesNothingOffTheBattlefield(t *testing.T) {
	g := newActiveGame(t)
	g.WithWriteLock(func() {
		if g.exertLocked(uuid.New(), g.Seats[0].ID, uuid.Nil) {
			t.Error("an object that isn't on the battlefield can't be exerted")
		}
	})
	if len(g.TurnTally.Exerts) != 0 {
		t.Error("nothing recorded")
	}
}

// TestStagedExertAndExertsSurviveARestore is ADR 0130 test 12's
// round-trip half: a restore point with a staged exert pays it at the
// lock-in, and one with TurnTally.Exerts still answers ExertedThisTurn.
// The shape guard records the two keys (testdata/snapshot_shape).
func TestStagedExertAndExertsSurviveARestore(t *testing.T) {
	stubExert(t, nil)
	g := newActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := pushExerter(t, g, me)
	b := pushExerter(t, g, me)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: a, Target: opp.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare a: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if err := g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: b, Target: opp.ID, Exert: true}, DeclareAttackersParams{}); err != nil {
		t.Fatalf("declare b: %v", err)
	}

	snap, restored := roundTrip(t, g)
	if len(snap.TurnTally.Exerts) != 1 {
		t.Fatalf("captured Exerts = %+v", snap.TurnTally.Exerts)
	}
	if !findCard(restored, b).ExertOnAttack {
		t.Fatal("the staged exert survives the restore")
	}
	restored.WithWriteLock(func() {
		if !restored.ExertedThisTurn(a) {
			t.Error("the restored tally still says a was exerted this turn")
		}
	})
	lockInAttackDeclaration(t, restored)
	if len(exertEvents(restored, b)) != 1 || len(restored.TurnTally.Exerts) != 2 {
		t.Fatal("the restored staged exert is paid at the lock-in")
	}
	_, again := roundTrip(t, restored)
	if len(again.TurnTally.Exerts) != 2 {
		t.Error("Exerts round-trips")
	}
}
