package game

import (
	"testing"

	"github.com/google/uuid"
)

// attack_event_test.go — S22: EventAttack, the event behind
// "whenever ~ attacks" triggers. Covers the payload, the
// once-per-creature firing rate, the step it fires in, the case
// where it must NOT fire (a creature that stayed home), and the
// harvester wiring that turns it into a stack item.
//
// Since #859 the event is emitted by the declaration's LOCK-IN
// rather than by each click, so every test here declares and then
// locks in; attack_declaration_test.go covers the lock-in itself.

// attackEvents returns every EventAttack in the log, in emit order.
func attackEvents(g *Game) []Event {
	var out []Event
	for _, ev := range g.Events {
		if ev.Kind == EventAttack {
			out = append(out, ev)
		}
	}
	return out
}

// stepRecorder is a test Listener that records the turn step each
// EventAttack was emitted in. Registered before combat so the
// emit-time step is observed rather than inferred afterwards.
type stepRecorder struct {
	steps []Step
}

func (r *stepRecorder) OnEvent(g *Game, ev Event) {
	if ev.Kind == EventAttack {
		r.steps = append(r.steps, g.Turn.Step)
	}
}

func TestAttackEventCarriesAttackerControllerAndDefender(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)

	evs := attackEvents(g)
	if len(evs) != 1 {
		t.Fatalf("attack events = %d, want 1", len(evs))
	}
	ev := evs[0]
	if ev.CardID != attacker {
		t.Errorf("CardID = %s, want the attacking creature %s", ev.CardID, attacker)
	}
	if ev.Actor != g.Seats[0].ID {
		t.Errorf("Actor = %s, want the attacker's controller %s", ev.Actor, g.Seats[0].ID)
	}
	if ev.Target != g.Seats[1].ID {
		t.Errorf("Target = %s, want the defending player %s", ev.Target, g.Seats[1].ID)
	}
}

// One event per attacking creature — the EventDrawCard "fires per
// card" contract. A creature left at home produces none, which is
// what stops a "whenever a creature you control attacks" payoff
// from counting the whole battlefield.
func TestAttackEventFiresOncePerAttackerAndNotForStayAtHomes(t *testing.T) {
	g := newActiveGame(t)
	a := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	b := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	c := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	home := pushKeywordCreature(t, g, g.Seats[0], 2, 2)

	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, id := range []uuid.UUID{a, b, c} {
		if err := g.DeclareAttacker(id, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	lockInAttackDeclaration(t, g)

	evs := attackEvents(g)
	if len(evs) != 3 {
		t.Fatalf("attack events = %d, want 3 (one per attacker)", len(evs))
	}
	seen := map[uuid.UUID]int{}
	for _, ev := range evs {
		seen[ev.CardID]++
	}
	for _, id := range []uuid.UUID{a, b, c} {
		if seen[id] != 1 {
			t.Errorf("attacker %s fired %d events, want exactly 1", id, seen[id])
		}
	}
	if seen[home] != 0 {
		t.Errorf("a creature that never attacked fired %d attack events", seen[home])
	}
}

// The event belongs to the declare-attackers step: it is emitted at
// the declaration's lock-in, which is the first priority boundary of
// that step, so a trigger lands ahead of blockers. Walking the rest
// of combat adds nothing — end of combat clears AttackingTarget
// without re-announcing anything.
func TestAttackEventFiresInTheDeclareAttackersStep(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	rec := &stepRecorder{}
	g.RegisterListener(rec)

	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)
	if len(rec.steps) != 1 {
		t.Fatalf("recorded %d attack events, want 1", len(rec.steps))
	}
	if rec.steps[0] != StepDeclareAttackers {
		t.Errorf("attack event emitted in step %s, want %s", rec.steps[0], StepDeclareAttackers)
	}

	advanceIntoStep(t, g, StepEndCombat)
	if len(rec.steps) != 1 {
		t.Errorf("combat produced %d attack events after declaration, want no more", len(rec.steps)-1)
	}
}

// Outside its step DeclareAttacker rejects the call, and a rejected
// declaration must not leave an event behind.
func TestAttackEventNotEmittedOutsideDeclareAttackers(t *testing.T) {
	g := newActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepPrecombatMain)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != ErrWrongStep {
		t.Fatalf("DeclareAttacker in main phase: got %v, want ErrWrongStep", err)
	}
	if n := len(attackEvents(g)); n != 0 {
		t.Errorf("rejected declaration emitted %d attack events, want 0", n)
	}
}

// The sandbox lets a player re-point an already-attacking creature
// at a different defender before the declaration is complete. That
// is one decision being revised, not two attacks: ONE event, naming
// the defender the creature ends on (#859).
func TestRedeclaringAnAttackerRetargetsWithoutRefiring(t *testing.T) {
	g := newFourPlayerActiveGame(t)
	attacker := pushKeywordCreature(t, g, g.Seats[0], 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if err := g.DeclareAttacker(attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("re-DeclareAttacker: %v", err)
	}
	lockInAttackDeclaration(t, g)

	evs := attackEvents(g)
	if len(evs) != 1 {
		t.Fatalf("attack events = %d, want 1 (re-pointing is not a second attack)", len(evs))
	}
	if evs[0].Target != g.Seats[2].ID {
		t.Errorf("event Target = %s, want the FINAL defender %s", evs[0].Target, g.Seats[2].ID)
	}
	card := findCard(g, attacker)
	if card == nil || card.AttackingTarget != g.Seats[2].ID {
		t.Errorf("AttackingTarget did not follow the re-declaration to %s", g.Seats[2].ID)
	}
}

// The dispatcher wiring: a catalog card watching EventAttack gets
// Build called and the item drained onto the stack by the lock-in,
// inside the declare-attackers step. Landing later would put an
// attack trigger after blockers.
func TestAttackTriggerReachesTheStackAtDeclarationTime(t *testing.T) {
	g := newActiveGame(t)
	owner := g.Seats[0]
	const oracle = "test-attack-oracle"
	cardID := uuid.New()
	g.Battlefield.PushTop(Card{
		InstanceID: cardID,
		Name:       "Test Attacker",
		OracleID:   oracle,
		TypeLine:   "Creature — Test",
		Power:      2,
		Toughness:  2,
		Owner:      owner.ID,
		Controller: owner.ID,
	})
	bystander := pushKeywordCreature(t, g, owner, 2, 2)

	var sawDefender uuid.UUID
	var buildCalls int
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
				buildCalls++
				sawDefender = ev.Target
				return NewTriggeredItem(source, "Test attack trigger")
			},
		}}
	})

	advanceIntoStep(t, g, StepDeclareAttackers)
	// The bystander attacks in the same declaration and must not fire
	// the ability — AppliesTo keys on the source's own instance ID —
	// so one declaration is worth exactly one Build call.
	if err := g.DeclareAttacker(bystander, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker bystander: %v", err)
	}
	if err := g.DeclareAttacker(cardID, g.Seats[1].ID); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	if buildCalls != 0 {
		t.Fatalf("Build ran before the declaration was locked in (%d calls)", buildCalls)
	}
	// Through the production path: the priority wrap is the lock-in.
	lockInAttacksByPriorityWrap(t, g)
	if buildCalls != 1 {
		t.Fatalf("Build calls = %d, want 1", buildCalls)
	}
	if sawDefender != g.Seats[1].ID {
		t.Errorf("Build saw defender %s, want %s", sawDefender, g.Seats[1].ID)
	}
	if len(g.PendingTriggers) != 0 {
		t.Errorf("trigger stranded in PendingTriggers: %d left undrained", len(g.PendingTriggers))
	}
	var onStack bool
	for _, item := range g.StackMeta {
		if item != nil && item.SourceCardID == cardID {
			onStack = true
		}
	}
	if !onStack {
		t.Errorf("attack trigger did not reach the stack during the declare-attackers step")
	}
	if g.Turn.Step != StepDeclareAttackers {
		t.Errorf("step moved to %s while declaring attackers", g.Turn.Step)
	}
}
