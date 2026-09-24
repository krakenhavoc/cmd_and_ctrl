package effects

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// block_declaration_test.go — #830 at the card level. The block
// declaration is announced once, when it is locked in (CR 509.1), so
// "becomes blocked" (CR 509.1h, afflict CR 702.130) fires once per
// blocked ATTACKER of the final assignment and "whenever this
// creature blocks" (CR 509.3a) once per blocker.
//
// The bug these pin: a blocker pointed at an attacker and then
// re-pointed elsewhere left the first attacker's trigger standing,
// so Cyberman Patrol afflicted for an attacker that ended the
// declaration unblocked.

// bdBlockEvents returns the logged events of `kind` naming `card` in
// CardID — the blocker for EventBlock, the attacker for
// EventBecomesBlocked.
func bdBlockEvents(g *game.Game, kind game.EventKind, card uuid.UUID) []game.Event {
	var out []game.Event
	for _, ev := range g.Events {
		if ev.Kind == kind && ev.CardID == card {
			out = append(out, ev)
		}
	}
	return out
}

// TestB26CybermanPatrolDoesNotAfflictAnAttackerTheBlockerLeft is the
// issue's own repro. The Patrol is blocked, the blocker is re-pointed
// at a second attacker, and the Patrol ends the declaration
// unblocked: no afflict, and the defender loses only the Patrol's
// combat damage.
func TestB26CybermanPatrolDoesNotAfflictAnAttackerTheBlockerLeft(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := b12Push(g, me.ID, "Cyberman Patrol", "Artifact Creature — Cyberman", b26CybermanPatrolOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, patrol, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life

	if err := g.DeclareBlocker(wall, patrol); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	// The re-point. The Bear is not an artifact creature, so the
	// final declaration produces no afflict at all.
	if err := g.DeclareBlocker(wall, bear); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInBlocks(t, g)

	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, patrol); n != 0 {
		t.Fatalf("the Patrol ends the declaration unblocked — no afflict: %d triggers", n)
	}
	if n := len(bdBlockEvents(g, game.EventBecomesBlocked, patrol)); n != 0 {
		t.Errorf("the attacker the blocker left never became blocked: %d events", n)
	}
	if n := len(bdBlockEvents(g, game.EventBecomesBlocked, bear)); n != 1 {
		t.Errorf("the attacker it was re-pointed to becomes blocked once: %d events", n)
	}
	advanceTo(t, g, game.StepCombatDamage)
	passPriorityAroundTable(t, g)
	if opp.Life != life-2 {
		t.Errorf("the unblocked Patrol's 2 damage and no afflict 3: %d to %d", life, opp.Life)
	}
}

// TestB18GrazilaxxBecomesBlockedFollowsARepointedBlocker — the same
// re-point seen through "whenever a creature you control becomes
// blocked": one trigger, and it names the attacker the blocker ended
// on.
func TestB18GrazilaxxBecomesBlockedFollowsARepointedBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Grazilaxx, Illithid Scholar", "Legendary Creature — Horror", b18GrazilaxxOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepDeclareBlockers)

	if err := g.DeclareBlocker(wall, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.DeclareBlocker(wall, b); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInBlocks(t, g)
	if n := b18TriggerPromptCount(g, me.ID); n != 1 {
		t.Fatalf("one blocked attacker is one prompt, got %d", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(b) {
		t.Error("the trigger is for the attacker the blocker ended on")
	}
	if me.Hand.Contains(a) {
		t.Error("the attacker the blocker left never became blocked")
	}
}

// TestB18GrazilaxxRepointedBackIsOneBecomesBlocked — A, B, then back
// to A. Three clicks, one declaration, one trigger, for A.
func TestB18GrazilaxxRepointedBackIsOneBecomesBlocked(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	b12Push(g, me.ID, "Grazilaxx, Illithid Scholar", "Legendary Creature — Horror", b18GrazilaxxOracle, 3, 2)
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepDeclareBlockers)

	for _, target := range []uuid.UUID{a, b, a} {
		if err := g.DeclareBlocker(wall, target); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	lockInBlocks(t, g)
	if n := b18TriggerPromptCount(g, me.ID); n != 1 {
		t.Fatalf("a round trip is one declaration and one prompt, got %d", n)
	}
	answerLatestTriggerPrompt(t, g, me.ID, true)
	passPriorityAroundTable(t, g)
	if !me.Hand.Contains(a) || me.Hand.Contains(b) {
		t.Error("the trigger is for A, the attacker the blocker ended on")
	}
}

// TestDoubleBlockIsOneAfflictAndOneBlockTriggerPerBlocker — CR 506.4
// blocks the attacker once however many creatures block it, while
// each blocker's own "whenever this creature blocks" fires
// (CR 509.3a).
func TestDoubleBlockIsOneAfflictAndOneBlockTriggerPerBlocker(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := b12Push(g, me.ID, "Cyberman Patrol", "Artifact Creature — Cyberman", b26CybermanPatrolOracle, 2, 2)
	h1 := b12Push(g, opp.ID, "Savvy Hunter", "Creature — Human Warrior", b30SavvyHunterOracle, 3, 3)
	h2 := b12Push(g, opp.ID, "Savvy Hunter", "Creature — Human Warrior", b30SavvyHunterOracle, 3, 3)
	declareAttack(t, g, opp.ID, patrol)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life

	for _, b := range []uuid.UUID{h1, h2} {
		if err := g.DeclareBlocker(b, patrol); err != nil {
			t.Fatalf("DeclareBlocker: %v", err)
		}
	}
	lockInBlocks(t, g)

	if n := len(bdBlockEvents(g, game.EventBecomesBlocked, patrol)); n != 1 {
		t.Errorf("a double block is one 'becomes blocked' (CR 506.4): %d events", n)
	}
	if n := len(bdBlockEvents(g, game.EventBlock, h1)) + len(bdBlockEvents(g, game.EventBlock, h2)); n != 2 {
		t.Errorf("each blocker blocks: %d block events, want 2", n)
	}
	// Two simultaneous triggers under one controller are ordered by
	// that controller (CR 603.3b) before anything reaches the stack.
	answerTriggerOrderInOfferedOrder(t, g)
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, opp.ID, "Food"); got != 2 {
		t.Errorf("one Food per blocking Hunter: %d", got)
	}
	if opp.Life != life-3 {
		t.Errorf("one afflict 3 for the double block, not two: %d to %d", life, opp.Life)
	}
}

// TestB30SavvyHunterBlocksOnceWhenRepointed — "whenever this creature
// blocks" fires once for a blocker that was moved between attackers
// before the declaration was complete, and the block it made names
// the attacker it ended on.
func TestB30SavvyHunterBlocksOnceWhenRepointed(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	a := b12Creature(g, me.ID, "Bear A", "Creature — Bear", 2, 2)
	b := b12Creature(g, me.ID, "Bear B", "Creature — Bear", 2, 2)
	hunter := b12Push(g, opp.ID, "Savvy Hunter", "Creature — Human Warrior", b30SavvyHunterOracle, 3, 3)
	declareAttack(t, g, opp.ID, a, b)
	advanceTo(t, g, game.StepDeclareBlockers)

	if err := g.DeclareBlocker(hunter, a); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if err := g.DeclareBlocker(hunter, b); err != nil {
		t.Fatalf("re-point: %v", err)
	}
	lockInBlocks(t, g)

	blocks := bdBlockEvents(g, game.EventBlock, hunter)
	if len(blocks) != 1 {
		t.Fatalf("a re-pointed blocker blocks once: %d block events", len(blocks))
	}
	if blocks[0].Target != b {
		t.Error("the block names the attacker the blocker ended on")
	}
	if n := len(bdBlockEvents(g, game.EventBecomesBlocked, a)); n != 0 {
		t.Errorf("the attacker it left never became blocked: %d events", n)
	}
	passPriorityAroundTable(t, g)
	if got := b30TokensNamed(g, opp.ID, "Food"); got != 1 {
		t.Errorf("one block is one Food: %d", got)
	}
}

// --- #492: block triggers reach the stack before combat damage -----
//
// The report: DeclareBlocker ran no state checks, so a "becomes
// blocked" trigger sat on PendingTriggers until the step advanced —
// after combat damage. #857 moved the announcement to the block
// declaration's LOCK-IN (commitBlockDeclarationLocked), which runs at
// the first priority boundary INSIDE declare_blockers, so the trigger
// is on the stack and resolves there under ordinary priority play.
// These two tests are that fix pinned, and the one case it does not
// cover.

// TestB492BlockTriggersResolveBeforeCombatDamage is the issue's own
// shape (CR 509.1h / 510): a blocked Cyberman Patrol's afflict 3 is
// applied in the declare-blockers step, ahead of the unblocked
// attacker's combat damage.
func TestB492BlockTriggersResolveBeforeCombatDamage(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := b12Push(g, me.ID, "Cyberman Patrol", "Artifact Creature — Cyberman", b26CybermanPatrolOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, patrol, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life

	if err := g.DeclareBlocker(wall, patrol); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	if n := len(g.PendingTriggers) + triggersOnStackFrom(g, patrol); n != 0 {
		t.Fatalf("the click stages the pairing and announces nothing: %d triggers", n)
	}
	lockInBlocks(t, g)
	if triggerOnStack(g, patrol) == nil && len(g.PendingTriggers) == 0 {
		t.Fatal("the lock-in announces the block and harvests the afflict")
	}
	passPriorityAroundTable(t, g)

	if g.Turn.Step != game.StepDeclareBlockers {
		t.Fatalf("the afflict resolved in %s, want declare_blockers", g.Turn.Step)
	}
	if opp.Life != life-3 {
		t.Fatalf("afflict 3 before any damage: %d to %d", life, opp.Life)
	}
	advanceTo(t, g, game.StepCombatDamage)
	if opp.Life != life-3-2 {
		t.Errorf("then the unblocked Bear's 2: %d, want %d", opp.Life, life-5)
	}
}

// TestB492AdvanceStepOutOfDeclareBlockersResolvesTheAfflictFirst is
// the same probe with the answer #914 gave it. It used to pin the
// opposite — the skip-ahead button walked the cursor into combat
// damage with the afflict still on the stack, so the Bear's 2 landed
// before the 3 — and that was never a property of block triggers:
// AdvanceStep walked past ANYTHING on the stack, including a trigger
// an earlier priority pass had announced.
//
// CR 117.4 is the rule it broke: a step ends only once every player
// has passed in succession with the stack EMPTY. advance_step now
// passes priority until the step ends, so what the step owes resolves
// inside it. The afflict is a declare-blockers trigger, it resolves
// there, and the 3 comes before the 2 — the same order ordinary
// priority play has produced since #857, which is why Cyberman Patrol
// no longer declares the caveat.
func TestB492AdvanceStepOutOfDeclareBlockersResolvesTheAfflictFirst(t *testing.T) {
	g := newCatalogGame(t)
	me, opp := g.Seats[0], g.Seats[1]
	patrol := b12Push(g, me.ID, "Cyberman Patrol", "Artifact Creature — Cyberman", b26CybermanPatrolOracle, 2, 2)
	bear := b12Creature(g, me.ID, "My Bear", "Creature — Bear", 2, 2)
	wall := b12Creature(g, opp.ID, "Wall", "Creature — Wall", 0, 4)
	declareAttack(t, g, opp.ID, patrol, bear)
	advanceTo(t, g, game.StepDeclareBlockers)
	life := opp.Life

	if err := g.DeclareBlocker(wall, patrol); err != nil {
		t.Fatalf("DeclareBlocker: %v", err)
	}
	advanceTo(t, g, game.StepCombatDamage)

	if triggerOnStack(g, patrol) != nil {
		t.Fatal("the afflict is still on the stack — the step ended owing it")
	}
	if opp.Life != life-3-2 {
		t.Fatalf("afflict 3 in declare_blockers, then the Bear's 2: %d to %d, want %d", life, opp.Life, life-5)
	}
	if spec, _ := Lookup(b26CybermanPatrolOracle); spec.Completeness != CompletenessFull {
		t.Error("the skip-ahead ordering was the card's last declared gap")
	}
}
