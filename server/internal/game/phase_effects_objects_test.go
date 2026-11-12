package game

import (
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// phase_effects_objects_test.go pins #5685's engine half (ADR 0188
// Decision 14): a requirement and a restriction bound to one future
// combat phase over creatures named as the spell resolves. The card is
// in cards/effects/illusionists_gambit_test.go.

// bindThoseCreatures binds Illusionist's Gambit's two rules to `phase`
// over `ids`, protecting `protected`, and reports whether the engine took
// it.
func bindThoseCreatures(t *testing.T, g *Game, phase int, protected uuid.UUID, ids ...uuid.UUID) bool {
	t.Helper()
	ok := false
	g.WithWriteLock(func() {
		ok = g.DuringCombatPhaseObjectsEffectForEffect(phase, uuid.Nil, ids, protected,
			[]Mod{AddAttackRequirementMod(uuid.Nil), CantAttackPlayerMod(protected)}, "test — those creatures")
	})
	return ok
}

// boundThatCombat reports whether the creature carries the requirement
// and the restriction right now.
func boundThatCombat(t *testing.T, g *Game, id uuid.UUID) (required, restricted bool) {
	t.Helper()
	ch := scopedEffectChar(t, g, id)
	return len(ch.AttackRequirements) > 0, len(ch.AttackTargetRestrictions) > 0
}

// TestPhaseObjectsEffectAppliesOnlyInItsCombat: bound to the combat it
// added, the rule is not in force in the combat that made it nor in a
// newer added combat that runs first (CR 500.8). In its own combat a
// named creature must attack (CR 508.1d) and can't attack the protected
// player (CR 508.1c), a creature that was not named is free, and the
// rule ends with that combat.
func TestPhaseObjectsEffectAppliesOnlyInItsCombat(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	caster := g.Seats[(g.Turn.ActiveSeat+1)%3]
	other := g.Seats[(g.Turn.ActiveSeat+2)%3]
	advanceIntoStep(t, g, StepBeginCombat)
	named := pushCombatant(t, g, me, "Named Bear", 2, 2)
	free := pushCombatant(t, g, me, "Free Bear", 2, 2)

	gambits := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if len(gambits) != 1 || !bindThoseCreatures(t, g, gambits[0], caster.ID, named) {
		t.Fatalf("setup: ids %v, or the binding was refused", gambits)
	}
	if len(g.PhaseEffects) != 1 || scopedEffectCount(g) != 0 {
		t.Fatalf("waiting %d, registered %d; want one waiting and none registered", len(g.PhaseEffects), scopedEffectCount(g))
	}
	if req, res := boundThatCombat(t, g, named); req || res {
		t.Fatal("the rule is in force in the combat that created it")
	}

	newer := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	advanceToPhase(t, g, newer[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	if req, res := boundThatCombat(t, g, named); req || res {
		t.Fatal("the rule is in force in a newer added combat")
	}
	if err := g.DeclareAttacker(named, caster.ID); err != nil {
		t.Fatalf("in the newer combat the named creature could not attack the protected player: %v", err)
	}
	if len(g.PhaseEffects) != 1 {
		t.Fatalf("the waiting rule was spent on the wrong combat: %+v", g.PhaseEffects)
	}

	advanceToPhase(t, g, gambits[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	if len(g.PhaseEffects) != 0 || scopedEffectCount(g) != 1 {
		t.Fatalf("in its combat: waiting %d, registered %d; want none waiting and one registered",
			len(g.PhaseEffects), scopedEffectCount(g))
	}
	g.WithWriteLock(func() { _ = g.UntapTargetForEffect(named) })
	if req, res := boundThatCombat(t, g, named); !req || !res {
		t.Fatalf("named creature in its combat: required %v, restricted %v; want both", req, res)
	}
	if req, res := boundThatCombat(t, g, free); req || res {
		t.Errorf("a creature that was not named: required %v, restricted %v", req, res)
	}
	if err := g.PassPriority(); !errors.Is(err, ErrAttackRequirement) {
		t.Errorf("pass with the named creature home = %v, want ErrAttackRequirement", err)
	}
	if err := g.DeclareAttacker(named, caster.ID); err == nil {
		t.Error("the named creature attacked the protected player")
	}
	if err := g.DeclareAttacker(named, other.ID); err != nil {
		t.Errorf("the named creature could not attack another opponent: %v", err)
	}
	if err := g.DeclareAttacker(free, caster.ID); err != nil {
		t.Errorf("a creature that was not named could not attack the protected player: %v", err)
	}

	advanceIntoStep(t, g, StepPostcombatMain)
	if req, res := boundThatCombat(t, g, named); req || res || scopedEffectCount(g) != 0 {
		t.Error("the rule outlived its combat")
	}
}

// TestPhaseObjectsEffectDropsAMemberThatChangedController: a named
// creature whose controller changes while the rule waits is dropped for
// good, even when its first controller has it back before that combat
// begins. The other member stays.
func TestPhaseObjectsEffectDropsAMemberThatChangedController(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	caster := g.Seats[(g.Turn.ActiveSeat+1)%3]
	advanceIntoStep(t, g, StepBeginCombat)
	stolen := pushCombatant(t, g, me, "Stolen Bear", 2, 2, "haste")
	stays := pushCombatant(t, g, me, "Staying Bear", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if !bindThoseCreatures(t, g, ids[0], caster.ID, stolen, stays) {
		t.Fatal("setup: the binding was refused")
	}
	// Stolen until the end of this combat, so it is home again before the
	// added combat begins.
	g.WithWriteLock(func() {
		d, ok := g.UntilEndOfCombatDuration()
		if !ok || !g.GainControlForEffect(uuid.New(), stolen, caster.ID, d, "test — steal") {
			t.Fatal("setup: the steal was refused")
		}
	})
	g.ReadSnapshot(func() {})
	if got := g.PhaseEffects; len(got) != 1 || len(got[0].Effect.Affected) != 1 || got[0].Effect.Affected[0].ID != stays {
		t.Fatalf("after the control change the waiting rule names %+v, want only the staying creature", got)
	}

	advanceToPhase(t, g, ids[0])
	advanceIntoStep(t, g, StepDeclareAttackers)
	if c, _ := g.LookupCardForEffect(stolen); c.Controller != me.ID {
		t.Fatalf("setup: the stolen creature did not come home (controller %s)", c.Controller)
	}
	if req, res := boundThatCombat(t, g, stolen); req || res {
		t.Errorf("the creature that changed controller: required %v, restricted %v; want neither", req, res)
	}
	if req, res := boundThatCombat(t, g, stays); !req || !res {
		t.Errorf("the creature that stayed: required %v, restricted %v; want both", req, res)
	}
	if err := g.DeclareAttacker(stolen, caster.ID); err != nil {
		t.Errorf("the creature that changed controller could not attack the protected player: %v", err)
	}
}

// TestPhaseObjectsEffectDropsAMemberUnderAnotherControllerAsCombatBegins:
// registration keeps only the members that are still that object under
// the controller they were pinned with, and a record left with none is
// spent without being registered.
func TestPhaseObjectsEffectDropsAMemberUnderAnotherControllerAsCombatBegins(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	caster := g.Seats[(g.Turn.ActiveSeat+1)%3]
	advanceIntoStep(t, g, StepBeginCombat)
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if !bindThoseCreatures(t, g, ids[0], caster.ID, bear) {
		t.Fatal("setup: the binding was refused")
	}
	// As a file written before the control change was noticed would be:
	// the member is pinned under a player who no longer controls it.
	g.PhaseEffects[0].Effect.Affected[0].Controller = caster.ID
	advanceToPhase(t, g, ids[0])
	if len(g.PhaseEffects) != 0 || scopedEffectCount(g) != 0 {
		t.Errorf("waiting %d, registered %d; want the record spent and nothing registered",
			len(g.PhaseEffects), scopedEffectCount(g))
	}
}

// TestPhaseObjectsEffectRefusals: nothing on the battlefield to name, no
// mods, and a phase that is not a combat ahead write nothing; a mod that
// changes a characteristic is a programming error.
func TestPhaseObjectsEffectRefusals(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	advanceIntoStep(t, g, StepBeginCombat)
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if bindThoseCreatures(t, g, ids[0], me.ID, uuid.New()) {
		t.Error("a set naming nothing on the battlefield was bound")
	}
	if bindThoseCreatures(t, g, g.Turn.PhaseID, me.ID, bear) {
		t.Error("the combat in progress was bound")
	}
	g.WithWriteLock(func() {
		if g.DuringCombatPhaseObjectsEffectForEffect(ids[0], uuid.Nil, []uuid.UUID{bear}, me.ID, nil, "test") {
			t.Error("a rule with no mods was bound")
		}
	})
	if len(g.PhaseEffects) != 0 {
		t.Errorf("a refusal wrote %+v", g.PhaseEffects)
	}
	defer func() {
		if s, _ := recover().(string); !strings.Contains(s, "carries only an attack requirement or restriction") {
			t.Errorf("want a registration panic for a characteristic mod, got %q", s)
		}
	}()
	g.WithWriteLock(func() {
		g.DuringCombatPhaseObjectsEffectForEffect(ids[0], uuid.Nil, []uuid.UUID{bear}, me.ID, []Mod{ModifyPTMod(1, 1)}, "test")
	})
}

// TestPhaseObjectsEffectCloneAndSnapshot: the waiting rule clones without
// aliasing its members, comes back from a restore point with them, and
// is registered in its combat of the restored game.
func TestPhaseObjectsEffectCloneAndSnapshot(t *testing.T) {
	g := newActiveGameWithSeats(t, 3)
	me := g.Seats[g.Turn.ActiveSeat]
	caster := g.Seats[(g.Turn.ActiveSeat+1)%3]
	advanceIntoStep(t, g, StepBeginCombat)
	bear := pushCombatant(t, g, me, "Bear", 2, 2)
	ids := addPhases(t, g, AnchorThisPhase, PhaseKindCombat)
	if !bindThoseCreatures(t, g, ids[0], caster.ID, bear) {
		t.Fatal("setup: the binding was refused")
	}

	cl := g.Clone()
	cl.PhaseEffects[0].Effect.Affected[0].ID = uuid.New()
	if g.PhaseEffects[0].Effect.Affected[0].ID != bear {
		t.Error("the clone aliases the live rule's members")
	}

	snap, restored := roundTrip(t, g)
	if len(snap.PhaseEffects) != 1 || len(restored.PhaseEffects) != 1 {
		t.Fatalf("snapshot %d, restored %d waiting rules; want 1 and 1", len(snap.PhaseEffects), len(restored.PhaseEffects))
	}
	got := restored.PhaseEffects[0]
	if got.PhaseID != ids[0] || got.Effect.Scope != ScopeNone || len(got.Effect.Affected) != 1 ||
		got.Effect.Affected[0].ID != bear || got.Effect.Affected[0].Controller != me.ID {
		t.Errorf("restore lost the binding: %+v", got)
	}
	advanceToPhase(t, restored, ids[0])
	if req, res := boundThatCombat(t, restored, bear); !req || !res {
		t.Errorf("restored game, in its combat: required %v, restricted %v; want both", req, res)
	}
	if err := func() error {
		advanceIntoStep(t, restored, StepDeclareAttackers)
		return restored.DeclareAttacker(bear, caster.ID)
	}(); err == nil {
		t.Error("after a restore the named creature attacked the protected player")
	}
}
