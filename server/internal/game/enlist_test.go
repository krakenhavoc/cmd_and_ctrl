package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// enlist_test.go — the engine half of #5658: enlist (CR 702.154) as a
// canonical keyword whose optional cost to attack is chosen with the
// declaration and paid at the lock-in, and whose linked trigger is
// derived from the ability list. The cards are in cards/effects.

// enlisterForTest is a 2/2 with enlist that can attack this turn.
func enlisterForTest(g *Game, owner uuid.UUID) uuid.UUID {
	return meleeAttackerForTest(g, owner, KeywordEnlist)
}

// helperForTest is a creature of the given power that has been under
// owner's control since the turn began.
func helperForTest(g *Game, owner uuid.UUID, power int) uuid.UUID {
	id := pushScopedTestCreature(g, owner, power, power)
	findBattlefieldCard(g, id).SummonedThisTurn = false
	return id
}

func enlistEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventEnlist {
			out = append(out, ev)
		}
	}
	return out
}

func enlistItems(g *Game) []*StackItem {
	var out []*StackItem
	g.ReadSnapshot(func() {
		for _, it := range g.PendingTriggers {
			if it != nil && it.Body == enlistPumpKey {
				out = append(out, it)
			}
		}
		for _, it := range g.StackMeta {
			if it != nil && it.Body == enlistPumpKey {
				out = append(out, it)
			}
		}
	})
	return out
}

func resolveEnlistItem(t *testing.T, g *Game) {
	t.Helper()
	g.WithWriteLock(func() {
		g.drainPendingTriggersAPNAPLocked()
		g.resolveTopAbilityLocked()
	})
}

func declareEnlisting(g *Game, attacker, target, enlisted uuid.UUID) error {
	return g.DeclareAttackerDeclWith(AttackDeclaration{Attacker: attacker, Target: target, Enlist: enlisted}, DeclareAttackersParams{})
}

func TestEnlistIsACanonicalKeywordWithOneLinkedTrigger(t *testing.T) {
	if kw, ok := CanonicalKeyword("Enlist"); !ok || kw != KeywordEnlist {
		t.Errorf("CanonicalKeyword(\"Enlist\") = %q, %v", kw, ok)
	}
	printed := Card{Name: "Recruiter", TypeLine: "Creature — Human Soldier", Keywords: []string{KeywordEnlist}}
	trigs := TriggersForCard(printed)
	if len(trigs) != 1 || trigs[0].Keyword != KeywordEnlist {
		t.Fatalf("a printed enlist has triggers %+v, want the one linked trigger", trigs)
	}
	silenced := printed
	empty := Characteristic{AbilitiesRemoved: true}
	silenced.effective = &empty
	if n := len(TriggersForCard(silenced)); n != 0 {
		t.Errorf("a creature with no abilities has %d enlist triggers", n)
	}
}

// TestEnlistIsPaidAtTheLockIn: the verb stages the choice and taps
// nothing; the lock-in taps the enlisted creature and emits EventEnlist
// in the declaration's batch, ahead of the EventAttack (CR 508.1j
// before 508.1k); the linked trigger adds the enlisted creature's power
// (CR 702.154a).
func TestEnlistIsPaidAtTheLockIn(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)

	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare with enlist: %v", err)
	}
	if findCard(g, a).EnlistOnAttack != h {
		t.Fatal("the choice is staged on the attacker")
	}
	if findCard(g, h).Tapped || len(enlistEvents(g)) != 0 {
		t.Fatal("nothing is paid before the lock-in")
	}
	lockInAttackDeclaration(t, g)

	if findCard(g, a).EnlistOnAttack != uuid.Nil {
		t.Error("the staged choice is spent at the lock-in")
	}
	if !findCard(g, h).Tapped {
		t.Fatal("the enlisted creature is tapped to pay the cost")
	}
	if findCard(g, h).AttackingTarget != uuid.Nil {
		t.Error("the enlisted creature is not attacking")
	}
	evs := enlistEvents(g)
	attacks := attackDeclEvents(g, a)
	if len(evs) != 1 || len(attacks) != 1 {
		t.Fatalf("%d enlist events and %d attack events, want one each", len(evs), len(attacks))
	}
	ev, at := evs[0], attacks[0]
	if ev.Actor != me.ID || ev.CardID != a || ev.Source != a || ev.Target != h {
		t.Errorf("enlist event = %+v", ev)
	}
	if ev.Batch != at.Batch || ev.Seq >= at.Seq {
		t.Errorf("enlist (batch %d, seq %d) must share the attack's batch (%d) and precede it (seq %d)", ev.Batch, ev.Seq, at.Batch, at.Seq)
	}
	items := enlistItems(g)
	if len(items) != 1 || items[0].Controller != me.ID || items[0].Label != enlistLabel {
		t.Fatalf("enlist triggers = %+v, want one", items)
	}
	resolveEnlistItem(t, g)
	if ch := scopedEffectChar(t, g, a); ch.Power != 5 || ch.Toughness != 2 {
		t.Errorf("enlister is %d/%d, want 5/2 (+3/+0)", ch.Power, ch.Toughness)
	}
}

// TestEnlistReadsThePowerAsItResolves: X is read as the trigger
// resolves (CR 608.2h), from the creature's last-known information if
// it has left the battlefield.
func TestEnlistReadsThePowerAsItResolves(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 1)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare: %v", err)
	}
	lockInAttackDeclaration(t, g)
	// The enlisted creature grows with the trigger waiting, then dies.
	g.WithWriteLock(func() {
		g.RegisterScopedEffectForEffect(h, []AffectedObject{PinObject(h, findBattlefieldCard(g, h).EnteredBattlefieldAt)},
			[]Mod{ModifyPTMod(3, 3)}, g.UntilEndOfTurnDuration(), "test pump")
		g.RecomputeLayersIfStaleLocked()
	})
	if err := g.MoveCardByID(ZoneRef{Kind: ZoneBattlefield}, ZoneRef{Kind: ZoneGraveyard, Owner: me.ID}, h); err != nil {
		t.Fatalf("move the enlisted creature: %v", err)
	}
	resolveEnlistItem(t, g)
	if ch := scopedEffectChar(t, g, a); ch.Power != 6 || ch.Toughness != 2 {
		t.Errorf("enlister is %d/%d, want 6/2 (+4/+0, the enlisted creature's last-known power)", ch.Power, ch.Toughness)
	}
}

func TestEnlistGivesNothingToAnAttackerThatBecameANewObject(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare: %v", err)
	}
	lockInAttackDeclaration(t, g)
	g.WithWriteLock(func() { findBattlefieldCard(g, a).ObjectEpoch++ })
	resolveEnlistItem(t, g)
	if ch := scopedEffectChar(t, g, a); ch.Power != 2 {
		t.Errorf("a new object got the pump: power %d, want 2", ch.Power)
	}
}

// TestEnlistRefusals is CR 702.154a's description of the creature that
// may be tapped, and 702.154c's "can't enlist itself". A refusal leaves
// the board as it was.
func TestEnlistRefusals(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	plain := attackerForTest(g, me.ID)
	ok := helperForTest(g, me.ID, 2)
	tapped := helperForTest(g, me.ID, 2)
	findBattlefieldCard(g, tapped).Tapped = true
	sick := pushScopedTestCreature(g, me.ID, 2, 2)
	findBattlefieldCard(g, sick).SummonedThisTurn = true
	hasty := pushKeywordCreature(t, g, me, 2, 2, "haste")
	findCard(g, hasty).Keywords = []string{"haste"}
	findCard(g, hasty).SummonedThisTurn = true
	theirs := helperForTest(g, opp.ID, 2)
	land := NewCard("Forest", me.ID)
	land.TypeLine = "Basic Land — Forest"
	g.Battlefield.PushTop(land)
	advanceIntoStep(t, g, StepDeclareAttackers)

	cases := []struct {
		name               string
		attacker, enlisted uuid.UUID
	}{
		{"an attacker without enlist", plain, ok},
		{"itself", a, a},
		{"a tapped creature", a, tapped},
		{"a creature that came under your control this turn", a, sick},
		{"an opponent's creature", a, theirs},
		{"a noncreature permanent", a, land.InstanceID},
		{"something not on the battlefield", a, uuid.New()},
	}
	for _, tc := range cases {
		if err := declareEnlisting(g, tc.attacker, opp.ID, tc.enlisted); !errors.Is(err, ErrCantEnlist) {
			t.Errorf("%s: err = %v, want ErrCantEnlist", tc.name, err)
		}
		if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: tc.attacker, Target: opp.ID, Enlist: tc.enlisted}}); !errors.Is(err, ErrCantEnlist) {
			t.Errorf("%s, bulk: err = %v, want ErrCantEnlist", tc.name, err)
		}
		if c := findCard(g, tc.attacker); c.AttackingTarget != uuid.Nil || c.Tapped || c.EnlistOnAttack != uuid.Nil {
			t.Fatalf("%s: a refusal staged something: %+v", tc.name, c)
		}
	}
	g.ReadSnapshot(func() {
		got := g.EnlistCandidatesForEffect(findBattlefieldCard(g, a))
		want := map[uuid.UUID]bool{plain: true, ok: true, hasty: true}
		if len(got) != len(want) {
			t.Errorf("candidates = %v, want %v", got, want)
		}
		for _, id := range got {
			if !want[id] {
				t.Errorf("candidate %s is not enlistable", id)
			}
		}
		if got := g.EnlistCandidatesForEffect(findBattlefieldCard(g, plain)); got != nil {
			t.Errorf("a creature without enlist has candidates %v", got)
		}
	})
	// A creature with haste may be enlisted the turn it arrives.
	if err := declareEnlisting(g, a, opp.ID, hasty); err != nil {
		t.Fatalf("enlisting a creature with haste: %v", err)
	}
}

// TestEnlistedCreatureCantAlsoAttackOrBeEnlistedTwice: "that you didn't
// choose to attack with", and one creature is tapped once.
func TestEnlistedCreatureCantAlsoAttackOrBeEnlistedTwice(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	b := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)

	// In one bulk declaration.
	if _, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: a, Target: opp.ID, Enlist: h}, {Attacker: h, Target: opp.ID},
	}); !errors.Is(err, ErrCantEnlist) {
		t.Errorf("enlisting a creature of the same declaration: %v, want ErrCantEnlist", err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: a, Target: opp.ID, Enlist: h}, {Attacker: b, Target: opp.ID, Enlist: h},
	}); !errors.Is(err, ErrCantEnlist) {
		t.Errorf("two attackers enlisting one creature: %v, want ErrCantEnlist", err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{
		{Attacker: a, Target: opp.ID, Enlist: b}, {Attacker: b, Target: opp.ID, Enlist: a},
	}); !errors.Is(err, ErrCantEnlist) {
		t.Errorf("two attackers enlisting each other: %v, want ErrCantEnlist", err)
	}
	for _, id := range []uuid.UUID{a, b, h} {
		if c := findCard(g, id); c.AttackingTarget != uuid.Nil || c.Tapped {
			t.Fatalf("a refused declaration staged %s", c.Name)
		}
	}

	// A creature at a time.
	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare a: %v", err)
	}
	if err := g.DeclareAttacker(h, opp.ID); !errors.Is(err, ErrEnlistedCantAttack) {
		t.Errorf("declaring the enlisted creature: %v, want ErrEnlistedCantAttack", err)
	}
	if err := declareEnlisting(g, b, opp.ID, h); !errors.Is(err, ErrCantEnlist) {
		t.Errorf("a second attacker enlisting it: %v, want ErrCantEnlist", err)
	}
	// An attacker already staged can't be enlisted either.
	if err := declareEnlisting(g, b, opp.ID, a); !errors.Is(err, ErrCantEnlist) {
		t.Errorf("enlisting an attacker: %v, want ErrCantEnlist", err)
	}
	// A re-point keeps the staged choice.
	if err := g.DeclareAttacker(a, g.Seats[2].ID); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	if findCard(g, a).EnlistOnAttack != h {
		t.Error("a re-point drops the staged enlist")
	}
	lockInAttackDeclaration(t, g)
	if n := len(enlistEvents(g)); n != 1 {
		t.Fatalf("%d enlist events, want 1", n)
	}
	// Once declared, an attacker can't enlist later in combat.
	h2 := helperForTest(g, me.ID, 2)
	if err := declareEnlisting(g, a, opp.ID, h2); !errors.Is(err, ErrCantEnlist) {
		t.Errorf("enlisting after the declaration: %v, want ErrCantEnlist", err)
	}
}

// TestStagedEnlistIsNotPaidWhenTheDeclarationIsTakenBack: a staged
// attack cleared before the lock-in never tapped anything, and a cost
// that can no longer be paid at the lock-in isn't paid.
func TestStagedEnlistIsNotPaidWhenTheDeclarationIsTakenBack(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare: %v", err)
	}
	if err := g.ClearCombat(); err != nil {
		t.Fatalf("clear combat: %v", err)
	}
	if findCard(g, a).EnlistOnAttack != uuid.Nil || findCard(g, h).Tapped {
		t.Fatal("clearing combat leaves the enlist staged or paid")
	}
	lockInAttackDeclaration(t, g)
	if n := len(enlistEvents(g)); n != 0 {
		t.Fatalf("%d enlist events after the declaration was taken back", n)
	}

	// The enlisted creature becomes tapped before the lock-in: the
	// cost can't be paid, so nothing is enlisted and the attack stands.
	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare again: %v", err)
	}
	g.WithWriteLock(func() { findBattlefieldCard(g, h).Tapped = true })
	lockInAttackDeclaration(t, g)
	if n := len(enlistEvents(g)); n != 0 {
		t.Errorf("%d enlist events for a cost that couldn't be paid", n)
	}
	if n := len(attackDeclEvents(g, a)); n != 1 {
		t.Errorf("%d attack events, want 1", n)
	}
	if n := len(enlistItems(g)); n != 0 {
		t.Errorf("%d enlist triggers with nothing enlisted", n)
	}
}

// TestStagedEnlistAndItsTriggerSurviveARestore: a restore point written
// between the verb and the lock-in still pays the enlist, and one
// written with the trigger waiting still resolves it. The shape guard
// records the key (testdata/snapshot_shape).
func TestStagedEnlistAndItsTriggerSurviveARestore(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 3)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := declareEnlisting(g, a, opp.ID, h); err != nil {
		t.Fatalf("declare: %v", err)
	}

	_, staged := roundTrip(t, g)
	if findCard(staged, a).EnlistOnAttack != h {
		t.Fatal("the staged enlist survives the restore")
	}
	lockInAttackDeclaration(t, staged)
	if !findCard(staged, h).Tapped || len(enlistItems(staged)) != 1 {
		t.Fatal("the restored game pays the enlist at the lock-in")
	}

	_, waiting := roundTrip(t, staged)
	items := enlistItems(waiting)
	if len(items) != 1 || len(items[0].Params.Objects) != 2 || items[0].Params.Objects[1].ID != h {
		t.Fatalf("restored enlist triggers = %+v", items)
	}
	resolveEnlistItem(t, waiting)
	if ch := scopedEffectChar(t, waiting, a); ch.Power != 5 {
		t.Errorf("restored enlister has power %d, want 5", ch.Power)
	}
}

// TestTheDeclarerOfAChosenDeclarationCantEnlist: the cost is the active
// player's (declaration_control.go).
func TestTheDeclarerOfAChosenDeclarationCantEnlist(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	chooseDeclarations(t, g, 1, true, false, false)
	a := enlisterForTest(g, me.ID)
	h := helperForTest(g, me.ID, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := declareEnlisting(g, a, g.Seats[2].ID, h); !errors.Is(err, ErrDeclarerCantEnlist) {
		t.Errorf("the declarer's enlist: %v, want ErrDeclarerCantEnlist", err)
	}
	if _, err := g.DeclareAttackers([]AttackDeclaration{{Attacker: a, Target: g.Seats[2].ID, Enlist: h}}); !errors.Is(err, ErrDeclarerCantEnlist) {
		t.Errorf("the declarer's bulk enlist: %v, want ErrDeclarerCantEnlist", err)
	}
}
