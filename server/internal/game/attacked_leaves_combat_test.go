package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// attacked_leaves_combat_test.go — #1376, ADR 0045 Decision 36.
//
// CR 506.4 removes an attacked planeswalker or battle from combat when
// its controller changes or it phases out, without it leaving the
// battlefield. CR 506.4c: the creatures attacking it "continue to be
// attacking creature[s], although [they are] not attacking any player,
// planeswalker, or battle. [They] may be blocked. If [they are]
// unblocked, [they] will deal no combat damage." CR 802.2a: the player
// who may block is the one defending BEFORE the removal — which
// Game.attackDefenders (#1364, Decision 35) already holds.

// stealForTest hands control of `target` to `to` until end of turn and
// runs the layer pass that materialises it (CR 613.1b), as a resolving
// Act of Treason would.
func stealForTest(t *testing.T, g *Game, target, to uuid.UUID) {
	t.Helper()
	g.WithWriteLock(func() {
		if !g.GainControlForEffect(uuid.New(), target, to, g.UntilEndOfTurnDuration(), "test — steal mid-combat") {
			t.Fatalf("GainControlForEffect refused %v", target)
		}
		g.RecomputeLayersIfStaleLocked()
	})
	if c := findBattlefieldCard(g, target); c == nil || c.Controller != to {
		t.Fatalf("setup: %v did not change control", target)
	}
}

// wantAttackingNothing asserts `attacker` is still an announced
// attacker, attacking nothing.
func wantAttackingNothing(t *testing.T, g *Game, attacker uuid.UUID) {
	t.Helper()
	c := findBattlefieldCard(g, attacker)
	if c == nil {
		t.Fatalf("the attacker left the battlefield")
	}
	if c.AttackingTarget != AttackingNothing {
		t.Errorf("CR 506.4c: attacker's target = %v, want AttackingNothing — what it attacked was removed from combat", c.AttackingTarget)
	}
	if !g.announcedAttacks[attacker] {
		t.Errorf("CR 506.4c: the attacker stopped being an attacking creature")
	}
}

// lifeTotals snapshots every seat's life.
func lifeTotals(g *Game) map[uuid.UUID]int {
	out := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		out[p.ID] = p.Life
	}
	return out
}

func wantLifeUnchanged(t *testing.T, g *Game, before map[uuid.UUID]int) {
	t.Helper()
	for _, p := range g.Seats {
		if p.Life != before[p.ID] {
			t.Errorf("%s's life %d -> %d; an attacker attacking nothing deals no damage (CR 506.4c)", p.Name, before[p.ID], p.Life)
		}
	}
}

// The issue's repro: seat 0 attacks seat 3's planeswalker, then steals
// it. The walker leaves combat; the attacker attacks nothing. Seat 3 —
// the defender before the removal — may block it, and the thief's
// opponents may not; the blocked attacker still fights its blocker.
func TestStolenAttackedWalkerLeavesCombatAndItsDefenderStillBlocks(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Stolen Walker", 4)
	guard := pushCombatant(t, g, g.Seats[3], "Walker's Guard", 4, 4)
	bystander := pushCombatant(t, g, g.Seats[1], "Bystander", 4, 4)
	declaredAttackerAt(t, g, attacker, walker)

	stealForTest(t, g, walker, g.Seats[0].ID)
	wantAttackingNothing(t, g, attacker)
	advanceTo(t, g, StepDeclareBlockers)

	if !offeredBlocks(g, g.Seats[3].ID)[[2]uuid.UUID{guard, attacker}] {
		t.Errorf("the generator does not offer the walker's former controller the block (CR 802.2a)")
	}
	if len(offeredBlocks(g, g.Seats[1].ID)) != 0 {
		t.Errorf("the generator offers a bystander a block")
	}
	refusal := wantNotDefending(t, g.DeclareBlocker(bystander, attacker), bystander)
	if refusal.Defender != g.Seats[3].ID {
		t.Errorf("refusal's defender = %v, want seat 3 (the controller before the removal)", refusal.Defender)
	}
	if err := g.DeclareBlocker(guard, attacker); err != nil {
		t.Fatalf("the walker's former controller blocks: %v", err)
	}

	life := lifeTotals(g)
	passUntilStep(t, g, StepEndCombat)
	if findBattlefieldCard(g, attacker) != nil {
		t.Errorf("the 4/4 blocker did not kill the 3/3 attacker it blocked")
	}
	wantLifeUnchanged(t, g, life)
	if got := findCard(g, walker).Counters[CounterLoyalty]; got != 4 {
		t.Errorf("walker loyalty = %d, want 4", got)
	}
}

// Unblocked, the attacker deals no damage — not to the walker, and
// not to its new controller. Stolen by a THIRD player here, so the
// live read (before #1376) resolved the attack to seat 1 and dealt
// the damage to the walker under seat 1's control.
func TestStolenAttackedWalkerTakesNoDamage(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Stolen Walker", 5)
	declaredAttackerAt(t, g, attacker, walker)

	stealForTest(t, g, walker, g.Seats[1].ID)
	wantAttackingNothing(t, g, attacker)

	life := lifeTotals(g)
	passUntilStep(t, g, StepEndCombat)
	wantLifeUnchanged(t, g, life)
	if got := findCard(g, walker).Counters[CounterLoyalty]; got != 5 {
		t.Errorf("walker loyalty = %d, want 5 — it was removed from combat (CR 506.4c)", got)
	}
}

// A battle that changes control is removed from combat too. Its
// PROTECTOR stays the defender (CR 310.9d through 802.2a), and an
// unblocked attacker takes no defense counters off it.
func TestStolenAttackedBattleLeavesCombat(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	blocked := pushCombatant(t, g, g.Seats[0], "Blocked", 2, 2)
	unblocked := pushCombatant(t, g, g.Seats[0], "Unblocked", 3, 3)
	battle := pushBattleForTest(g, g.Seats[1].ID, g.Seats[2].ID, "Stolen Siege", 5)
	controllers := pushCombatant(t, g, g.Seats[1], "Battle's Controller", 1, 4)
	protectors := pushCombatant(t, g, g.Seats[2], "Battle's Protector", 1, 4)
	thieves := pushCombatant(t, g, g.Seats[3], "Thief's Creature", 1, 4)
	advanceTo(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{blocked, unblocked} {
		if err := g.DeclareAttacker(a, battle); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	stealForTest(t, g, battle, g.Seats[3].ID)
	wantAttackingNothing(t, g, blocked)
	wantAttackingNothing(t, g, unblocked)
	advanceTo(t, g, StepDeclareBlockers)

	for _, b := range []uuid.UUID{controllers, thieves} {
		wantNotDefending(t, g.DeclareBlocker(b, blocked), b)
	}
	if err := g.DeclareBlocker(protectors, blocked); err != nil {
		t.Fatalf("the battle's protector blocks: %v", err)
	}

	life := lifeTotals(g)
	passUntilStep(t, g, StepEndCombat)
	wantLifeUnchanged(t, g, life)
	if got := findCard(g, battle).Counters[CounterDefense]; got != 5 {
		t.Errorf("battle defense = %d, want 5 — it was removed from combat (CR 506.4c)", got)
	}
}

// Phasing: CR 702.26b removes a phased-out permanent from combat. A
// walker that phases out and back in inside one combat does not come
// back ATTACKED — its absence from the battlefield slice stopped the
// damage while it was out, and the rewrite keeps it stopped once it
// returns. Its controller, who defended it, may still block.
func TestAttackedWalkerThatPhasesOutAndBackInStaysOutOfCombat(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	blocked := pushCombatant(t, g, g.Seats[0], "Blocked", 2, 2)
	unblocked := pushCombatant(t, g, g.Seats[0], "Unblocked", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Blinking Walker", 4)
	guard := pushCombatant(t, g, g.Seats[2], "Walker's Guard", 4, 4)
	lock := pushCombatant(t, g, g.Seats[1], "Oubliette Stand-in", 0, 1)
	advanceTo(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{blocked, unblocked} {
		if err := g.DeclareAttacker(a, walker); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })

	g.WithWriteLock(func() {
		if err := g.PhaseOutUntilLeavesForEffect(lock, lock, false, walker); err != nil {
			t.Fatalf("phase out: %v", err)
		}
	})
	removeFromBattlefieldForTest(t, g, lock, false)
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	if findBattlefieldCard(g, walker) == nil {
		t.Fatalf("setup: the walker did not phase back in")
	}
	wantAttackingNothing(t, g, blocked)
	wantAttackingNothing(t, g, unblocked)

	advanceTo(t, g, StepDeclareBlockers)
	if !offeredBlocks(g, g.Seats[2].ID)[[2]uuid.UUID{guard, blocked}] {
		t.Errorf("the generator does not offer the walker's controller the block")
	}
	if err := g.DeclareBlocker(guard, blocked); err != nil {
		t.Fatalf("the walker's controller blocks: %v", err)
	}

	life := lifeTotals(g)
	passUntilStep(t, g, StepEndCombat)
	wantLifeUnchanged(t, g, life)
	if got := findCard(g, walker).Counters[CounterLoyalty]; got != 4 {
		t.Errorf("walker loyalty = %d, want 4 — it phased out of combat (CR 702.26b, 506.4c)", got)
	}
	if findBattlefieldCard(g, blocked) != nil {
		t.Errorf("the 4/4 blocker did not kill the 2/2 attacker it blocked")
	}
}

// Regeneration is NOT a removal of the attacked side. CR 701.19a
// removes a regenerating permanent from combat only "if it's an
// attacking or blocking creature", so an attacked planeswalker that is
// also a creature (an animated Gideon) and regenerates stays attacked,
// and the creature attacking it still deals it damage.
func TestAttackedWalkerCreatureThatRegeneratesStaysAttacked(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Animated Walker", 5)
	g.WithWriteLock(func() {
		c := findBattlefieldCard(g, walker)
		c.TypeLine = "Legendary Planeswalker Creature — Test"
		c.Power, c.Toughness = 5, 5
	})
	declaredAttackerAt(t, g, attacker, walker)

	regenerate(t, g, walker)
	destroy(t, g, walker)
	if findBattlefieldCard(g, walker) == nil {
		t.Fatalf("setup: the walker did not regenerate")
	}
	if got := findBattlefieldCard(g, attacker).AttackingTarget; got != walker {
		t.Fatalf("a regenerating attacked walker left combat (attacker now attacking %v); CR 701.19a removes only an attacking or blocking creature", got)
	}
	passUntilStep(t, g, StepEndCombat)
	if got := findCard(g, walker).Counters[CounterLoyalty]; got != 2 {
		t.Errorf("walker loyalty = %d, want 2 — it was still attacked", got)
	}
}

// A creature only STAGED in the declare-attackers step is not in
// combat yet, and its declaration is still the active player's to
// change: a control change does not rewrite it.
func TestStagedAttackOnAStolenWalkerIsNotRewritten(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Walker", 4)
	advanceTo(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	stealForTest(t, g, walker, g.Seats[1].ID)
	if got := findBattlefieldCard(g, attacker).AttackingTarget; got != walker {
		t.Errorf("a staged declaration was rewritten to %v", got)
	}
}

// Undo and a persisted restore carry the rewrite: the creature is
// still attacking nothing, still blockable by the defender before the
// removal, and still deals no damage. Combat ending clears it.
func TestAttackingNothingRidesCloneAndRestore(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Stolen Walker", 4)
	guard := pushCombatant(t, g, g.Seats[3], "Guard", 1, 4)
	declaredAttackerAt(t, g, attacker, walker)
	stealForTest(t, g, walker, g.Seats[1].ID)
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
	// Undo is RestoreFrom a Clone taken earlier.
	undo := newActiveGameWithSeats(t, 4)
	undo.RestoreFrom(g.Clone())

	for name, h := range map[string]*Game{"clone": g.Clone(), "restore": restored, "undo": undo} {
		wantAttackingNothing(t, h, attacker)
		if err := h.DeclareBlocker(guard, attacker); err != nil {
			t.Errorf("%s: the walker's former controller cannot block: %v", name, err)
		}
		loyalty := findCard(h, walker).Counters[CounterLoyalty]
		life := lifeTotals(h)
		passUntilStep(t, h, StepEndCombat)
		wantLifeUnchanged(t, h, life)
		if got := findCard(h, walker).Counters[CounterLoyalty]; got != loyalty {
			t.Errorf("%s: walker loyalty %d -> %d", name, loyalty, got)
		}
	}

	passUntilStep(t, g, StepPostcombatMain)
	if c := findBattlefieldCard(g, attacker); c != nil && c.AttackingTarget != uuid.Nil {
		t.Errorf("combat ended and the attacker is still attacking %v", c.AttackingTarget)
	}
}
