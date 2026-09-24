package game

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
)

// attacked_loses_type_test.go — #1387, ADR 0045 Decision 37.
//
// CR 506.4: a permanent is removed from combat "if it is a planeswalker
// that stops being a planeswalker or a battle that stops being a
// battle". CR 506.4c: the creatures attacking it keep attacking, attack
// nothing, may be blocked by the player who was defending it (CR
// 802.2a), and deal no combat damage if unblocked. The removal is
// permanent: a type that comes back later in the same combat does not
// put the permanent back into combat.

// stripAttackableTypesForTest makes `target` lose the planeswalker and
// battle card types for as long as `lock` is on the battlefield — a
// layer-4 type removal from a floating continuous effect (CR 613.1d,
// 611.2b) — and runs the layer pass that applies it. Removing `lock`
// ends the effect and the type comes back on the next pass.
func stripAttackableTypesForTest(t *testing.T, g *Game, target, lock uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		d, ok := g.ForAsLongAsOnBattlefieldDuration(lock)
		if !ok {
			t.Fatalf("setup: the lock %v is not on the battlefield", lock)
		}
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer: Layer4Type,
			AppliesTo: func(c *Card, _ *Game, _ *Card) bool {
				return c.InstanceID == target
			},
			Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
				ch.Types = slices.DeleteFunc(slices.Clone(ch.Types), func(ty string) bool {
					return strings.EqualFold(ty, "Planeswalker") || strings.EqualFold(ty, "Battle")
				})
			},
		}, lock, "test — loses the planeswalker and battle types", d)
		g.RecomputeLayersIfStaleLocked()
	})
	if c := findBattlefieldCard(g, target); c == nil || c.IsPlaneswalker() || c.IsBattle() {
		t.Fatalf("setup: %v is still a planeswalker or battle", target)
	}
}

// releaseLockForTest removes `lock`, which ends the type removal, and
// runs the layer pass that brings the type back.
func releaseLockForTest(t *testing.T, g *Game, lock, target uuid.UUID) {
	t.Helper()
	removeFromBattlefieldForTest(t, g, lock, false)
	g.WithWriteLock(func() { g.RecomputeLayersIfStaleLocked() })
	if c := findBattlefieldCard(g, target); c == nil || !(c.IsPlaneswalker() || c.IsBattle()) {
		t.Fatalf("setup: %v did not get its type back", target)
	}
}

// The issue's repro. Seat 0 attacks seat 2's planeswalker with two
// creatures; the walker stops being a planeswalker. Both attackers keep
// attacking, attacking nothing. Seat 2 — the defender before the
// removal — blocks one and a bystander may not. The type then comes
// back before damage: the walker is a planeswalker again, but it left
// combat for good, so the unblocked attacker deals it nothing.
func TestAttackedWalkerThatStopsBeingOneLeavesCombatForGood(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	blocked := pushCombatant(t, g, g.Seats[0], "Blocked", 2, 2)
	unblocked := pushCombatant(t, g, g.Seats[0], "Unblocked", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Unmade Walker", 5)
	guard := pushCombatant(t, g, g.Seats[2], "Walker's Guard", 4, 4)
	bystander := pushCombatant(t, g, g.Seats[1], "Bystander", 4, 4)
	lock := pushCombatant(t, g, g.Seats[1], "Type-Stripper Stand-in", 0, 1)
	advanceTo(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{blocked, unblocked} {
		if err := g.DeclareAttacker(a, walker); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	stripAttackableTypesForTest(t, g, walker, lock)
	wantAttackingNothing(t, g, blocked)
	wantAttackingNothing(t, g, unblocked)
	advanceTo(t, g, StepDeclareBlockers)

	if !offeredBlocks(g, g.Seats[2].ID)[[2]uuid.UUID{guard, blocked}] {
		t.Errorf("the generator does not offer the walker's controller the block (CR 802.2a)")
	}
	if len(offeredBlocks(g, g.Seats[1].ID)) != 0 {
		t.Errorf("the generator offers a bystander a block")
	}
	refusal := wantNotDefending(t, g.DeclareBlocker(bystander, blocked), bystander)
	if refusal.Defender != g.Seats[2].ID {
		t.Errorf("refusal's defender = %v, want seat 2 (the walker's controller before the removal)", refusal.Defender)
	}
	if err := g.DeclareBlocker(guard, blocked); err != nil {
		t.Fatalf("the walker's controller blocks: %v", err)
	}

	releaseLockForTest(t, g, lock, walker)
	wantAttackingNothing(t, g, blocked)
	wantAttackingNothing(t, g, unblocked)

	life := lifeTotals(g)
	passUntilStep(t, g, StepEndCombat)
	wantLifeUnchanged(t, g, life)
	if got := findCard(g, walker).Counters[CounterLoyalty]; got != 5 {
		t.Errorf("walker loyalty = %d, want 5 — it stopped being a planeswalker and left combat for good (CR 506.4)", got)
	}
	if findBattlefieldCard(g, blocked) != nil {
		t.Errorf("the 4/4 blocker did not kill the 2/2 attacker it blocked")
	}
}

// A battle that stops being a battle leaves combat the same way: its
// PROTECTOR still blocks, and an unblocked attacker takes no defense
// counters off it once it is a battle again.
func TestAttackedBattleThatStopsBeingOneLeavesCombatForGood(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	blocked := pushCombatant(t, g, g.Seats[0], "Blocked", 2, 2)
	battle := pushBattleForTest(g, g.Seats[1].ID, g.Seats[2].ID, "Unmade Siege", 5)
	controllers := pushCombatant(t, g, g.Seats[1], "Battle's Controller", 1, 4)
	protectors := pushCombatant(t, g, g.Seats[2], "Battle's Protector", 1, 4)
	lock := pushCombatant(t, g, g.Seats[3], "Type-Stripper Stand-in", 0, 1)
	advanceTo(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{attacker, blocked} {
		if err := g.DeclareAttacker(a, battle); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	stripAttackableTypesForTest(t, g, battle, lock)
	wantAttackingNothing(t, g, attacker)
	releaseLockForTest(t, g, lock, battle)
	wantAttackingNothing(t, g, attacker)
	wantAttackingNothing(t, g, blocked)
	advanceTo(t, g, StepDeclareBlockers)

	wantNotDefending(t, g.DeclareBlocker(controllers, blocked), controllers)
	if err := g.DeclareBlocker(protectors, blocked); err != nil {
		t.Fatalf("the battle's protector blocks: %v", err)
	}
	life := lifeTotals(g)
	passUntilStep(t, g, StepEndCombat)
	wantLifeUnchanged(t, g, life)
	if got := findCard(g, battle).Counters[CounterDefense]; got != 5 {
		t.Errorf("battle defense = %d, want 5 — it stopped being a battle and left combat (CR 506.4)", got)
	}
}

// Only a LOSS of the type removes the permanent. A walker that becomes
// a creature as well stays a planeswalker, stays attacked, and takes
// the damage.
func TestAttackedWalkerThatGainsATypeStaysAttacked(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Animated Walker", 5)
	declaredAttackerAt(t, g, attacker, walker)

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer4Type,
			AppliesTo: func(c *Card, _ *Game, _ *Card) bool { return c.InstanceID == walker },
			Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
				ch.Types = append(slices.Clone(ch.Types), "Creature")
				ch.Power, ch.Toughness = 5, 5
			},
		}, walker, "test — becomes a 5/5 creature too", g.UntilEndOfTurnDuration())
		g.RecomputeLayersIfStaleLocked()
	})
	if c := findBattlefieldCard(g, walker); !c.IsCreature() || !c.IsPlaneswalker() {
		t.Fatalf("setup: the walker is not a planeswalker creature")
	}
	if got := findBattlefieldCard(g, attacker).AttackingTarget; got != walker {
		t.Fatalf("a walker that GAINED a type left combat (attacker now attacking %v)", got)
	}
	passUntilStep(t, g, StepEndCombat)
	if got := findCard(g, walker).Counters[CounterLoyalty]; got != 2 {
		t.Errorf("walker loyalty = %d, want 2 — it was still attacked", got)
	}
}

// The battle half of the same guard: a battle that becomes an artifact
// as well is still a battle, stays attacked, and loses the defense
// counters.
func TestAttackedBattleThatGainsATypeStaysAttacked(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	battle := pushBattleForTest(g, g.Seats[1].ID, g.Seats[2].ID, "Gilded Siege", 5)
	declaredAttackerAt(t, g, attacker, battle)

	g.WithWriteLock(func() {
		g.RegisterScopedStaticForEffect(StaticAbility{
			Layer:     Layer4Type,
			AppliesTo: func(c *Card, _ *Game, _ *Card) bool { return c.InstanceID == battle },
			Apply: func(ch *Characteristic, _ *Card, _ *Game, _ *Card) {
				ch.Types = append(slices.Clone(ch.Types), "Artifact")
			},
		}, battle, "test — becomes an artifact too", g.UntilEndOfTurnDuration())
		g.RecomputeLayersIfStaleLocked()
	})
	if c := findBattlefieldCard(g, battle); !c.IsArtifact() || !c.IsBattle() {
		t.Fatalf("setup: the battle is not an artifact battle")
	}
	if got := findBattlefieldCard(g, attacker).AttackingTarget; got != battle {
		t.Fatalf("a battle that GAINED a type left combat (attacker now attacking %v)", got)
	}
	passUntilStep(t, g, StepEndCombat)
	if got := findCard(g, battle).Counters[CounterDefense]; got != 2 {
		t.Errorf("battle defense = %d, want 2 — it was still attacked", got)
	}
}

// A creature only STAGED in the declare-attackers step is not in combat
// yet, so the type loss does not rewrite its declaration (Decision 36's
// rule, reached through the new door).
func TestStagedAttackOnAWalkerThatLosesItsTypeIsNotRewritten(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Walker", 4)
	lock := pushCombatant(t, g, g.Seats[1], "Type-Stripper Stand-in", 0, 1)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	stripAttackableTypesForTest(t, g, walker, lock)
	if got := findBattlefieldCard(g, attacker).AttackingTarget; got != walker {
		t.Errorf("a staged declaration was rewritten to %v", got)
	}
}

// Undo round-trips both ways. A clone taken BEFORE the type loss
// restores the original attack: the walker is attacked and takes the
// damage. A clone, undo and persisted restore taken AFTER the loss and
// the type's return carry the rewrite: still attacking nothing, still
// blockable by the walker's controller, still no damage.
func TestAttackedWalkerTypeLossRidesUndoAndRestore(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Unmade Walker", 5)
	guard := pushCombatant(t, g, g.Seats[3], "Guard", 1, 4)
	lock := pushCombatant(t, g, g.Seats[1], "Type-Stripper Stand-in", 0, 1)
	declaredAttackerAt(t, g, attacker, walker)
	before := g.Clone()

	stripAttackableTypesForTest(t, g, walker, lock)
	releaseLockForTest(t, g, lock, walker)
	wantAttackingNothing(t, g, attacker)

	// Undo to before the loss: the attack on the walker is back.
	rewound := newActiveGameWithSeats(t, 4)
	rewound.RestoreFrom(before)
	if got := findBattlefieldCard(rewound, attacker).AttackingTarget; got != walker {
		t.Fatalf("undo to before the type loss: attacker attacking %v, want the walker", got)
	}
	passUntilStep(t, rewound, StepEndCombat)
	if got := findCard(rewound, walker).Counters[CounterLoyalty]; got != 2 {
		t.Errorf("undo to before the type loss: walker loyalty = %d, want 2", got)
	}

	advanceTo(t, g, StepDeclareBlockers)
	raw, err := json.Marshal(g.CaptureSnapshot())
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}
	var snap GameSnapshot
	if err := json.Unmarshal(raw, &snap); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	restored, err := snap.Restore()
	if err != nil {
		t.Fatalf("restore: %v", err)
	}
	undo := newActiveGameWithSeats(t, 4)
	undo.RestoreFrom(g.Clone())

	for name, h := range map[string]*Game{"clone": g.Clone(), "restore": restored, "undo": undo} {
		wantAttackingNothing(t, h, attacker)
		if err := h.DeclareBlocker(guard, attacker); err != nil {
			t.Errorf("%s: the walker's controller cannot block: %v", name, err)
		}
		life := lifeTotals(h)
		passUntilStep(t, h, StepEndCombat)
		wantLifeUnchanged(t, h, life)
		if got := findCard(h, walker).Counters[CounterLoyalty]; got != 5 {
			t.Errorf("%s: walker loyalty = %d, want 5", name, got)
		}
	}
}
