package game

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

// attack_defender_lki_test.go — #1364, CR 506.4c: "If a creature is
// attacking a planeswalker or battle, removing that planeswalker or
// battle from combat doesn't remove that creature from combat. It
// continues to be an attacking creature, although it is not attacking
// any player, planeswalker, or battle. It may be blocked. If it is
// unblocked, it will deal no combat damage."
//
// CR 802.2a names who may block it: the player it was attacking, the
// controller of the planeswalker or the protector of the battle it was
// attacking, "before it was removed from combat". The engine records
// that player when the attack is pointed (Game.attackDefenders) and
// the block path falls back to it once the live target is gone.

// offeredBlocks is every (blocker, attacker) pair the ONE option
// generator offers `seat`.
func offeredBlocks(g *Game, seat uuid.UUID) map[[2]uuid.UUID]bool {
	out := map[[2]uuid.UUID]bool{}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, opt := range g.BlockOptionsLocked(seat, 4) {
			for _, d := range opt.Blocks {
				out[[2]uuid.UUID{d.Blocker, d.Attacker}] = true
			}
		}
	})
	return out
}

// removeFromBattlefieldForTest takes a permanent off the battlefield
// through the real exit door: destroy (to the graveyard) or exile.
func removeFromBattlefieldForTest(t *testing.T, g *Game, id uuid.UUID, exile bool) {
	t.Helper()
	var err error
	g.WithWriteLock(func() {
		if exile {
			err = g.ExileCardForEffect(id)
		} else {
			err = g.DestroyPermanentForEffect(id)
		}
	})
	if err != nil {
		t.Fatalf("removing %v: %v", id, err)
	}
	if findBattlefieldCard(g, id) != nil {
		t.Fatalf("setup: %v is still on the battlefield", id)
	}
}

// The issue's first shape: the planeswalker dies before blocks. Its
// former controller may still block the attacker — and kill it — and
// nobody else may; unblocked or not, the attacker deals no damage.
func TestAttackerWhoseWalkerDiedIsBlockedByItsFormerController(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Doomed Walker", 4)
	guard := pushCombatant(t, g, g.Seats[3], "Walker's Guard", 4, 4)
	bystander := pushCombatant(t, g, g.Seats[1], "Bystander", 4, 4)
	declaredAttackerAt(t, g, attacker, walker)

	removeFromBattlefieldForTest(t, g, walker, false)
	advanceTo(t, g, StepDeclareBlockers)

	if c := findCard(g, attacker); c.AttackingTarget != walker {
		t.Fatalf("CR 506.4c: the attacker left combat with its walker (attacking %v)", c.AttackingTarget)
	}
	if !offeredBlocks(g, g.Seats[3].ID)[[2]uuid.UUID{guard, attacker}] {
		t.Errorf("the generator does not offer the walker's former controller the block")
	}
	if len(offeredBlocks(g, g.Seats[1].ID)) != 0 {
		t.Errorf("the generator offers a bystander a block on an attack that was never theirs")
	}
	refusal := wantNotDefending(t, g.DeclareBlocker(bystander, attacker), bystander)
	if refusal.Defender != g.Seats[3].ID {
		t.Errorf("refusal's defender = %v, want seat 3 (the walker's controller)", refusal.Defender)
	}
	want := "Attacker is still attacking, though what it attacked is gone, so only P4 can block it."
	if got := refusal.Sentence(g.Seats[1].ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareBlocker(guard, attacker); err != nil {
		t.Fatalf("the walker's former controller blocks: %v", err)
	}

	life := g.Seats[3].Life
	passUntilStep(t, g, StepEndCombat)
	if findBattlefieldCard(g, attacker) != nil {
		t.Errorf("the 4/4 blocker did not kill the 3/3 attacker it blocked")
	}
	if g.Seats[3].Life != life {
		t.Errorf("seat 3 took damage from an attack on a walker that is gone: %d -> %d", life, g.Seats[3].Life)
	}
}

// Unblocked, it still deals its damage to nothing (CR 506.4c, 510.1b):
// the fallback belongs to blocking only.
func TestAttackerWhoseWalkerDiedDealsNoDamageUnblocked(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Doomed Walker", 4)
	declaredAttackerAt(t, g, attacker, walker)
	removeFromBattlefieldForTest(t, g, walker, false)

	life := map[uuid.UUID]int{}
	for _, p := range g.Seats {
		life[p.ID] = p.Life
	}
	passUntilStep(t, g, StepEndCombat)
	for _, p := range g.Seats {
		if p.Life != life[p.ID] {
			t.Errorf("%s's life %d -> %d; an attacker whose walker left deals no damage", p.Name, life[p.ID], p.Life)
		}
	}
}

// The second shape: a battle is exiled before blocks. Its PROTECTOR,
// not its controller, may block (CR 310.9d carried through 802.2a).
func TestAttackerWhoseBattleLeftIsBlockedByItsFormerProtector(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	battle := pushBattleForTest(g, g.Seats[1].ID, g.Seats[2].ID, "Vanishing Siege", 5)
	controllers := pushCombatant(t, g, g.Seats[1], "Battle's Controller", 1, 4)
	protectors := pushCombatant(t, g, g.Seats[2], "Battle's Protector", 1, 4)
	declaredAttackerAt(t, g, attacker, battle)

	removeFromBattlefieldForTest(t, g, battle, true)
	advanceTo(t, g, StepDeclareBlockers)

	if offeredBlocks(g, g.Seats[1].ID)[[2]uuid.UUID{controllers, attacker}] {
		t.Errorf("the generator offers the battle's controller the block")
	}
	if !offeredBlocks(g, g.Seats[2].ID)[[2]uuid.UUID{protectors, attacker}] {
		t.Errorf("the generator does not offer the battle's former protector the block")
	}
	wantNotDefending(t, g.DeclareBlocker(controllers, attacker), controllers)
	if err := g.DeclareBlocker(protectors, attacker); err != nil {
		t.Fatalf("the battle's former protector blocks: %v", err)
	}
}

// #1343's reselect updates the record: an attack moved from seat 1's
// walker onto seat 2's, whose walker then dies, is seat 2's to block.
func TestReselectUpdatesTheRecordedDefender(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Redirected", 3, 3)
	first := pushPlaneswalkerForTest(g, g.Seats[1].ID, "First Walker", 4)
	second := pushPlaneswalkerForTest(g, g.Seats[2].ID, "Second Walker", 4)
	oldGuard := pushCombatant(t, g, g.Seats[1], "First's Guard", 1, 4)
	newGuard := pushCombatant(t, g, g.Seats[2], "Second's Guard", 1, 4)
	declaredAttackerAt(t, g, attacker, first)
	if err := reselect(g, attacker, second); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}

	removeFromBattlefieldForTest(t, g, second, false)
	advanceTo(t, g, StepDeclareBlockers)

	if len(offeredBlocks(g, g.Seats[1].ID)) != 0 {
		t.Errorf("the generator offers the ORIGINAL walker's controller a block after the reselect")
	}
	if !offeredBlocks(g, g.Seats[2].ID)[[2]uuid.UUID{newGuard, attacker}] {
		t.Errorf("the generator does not offer the reselected walker's controller the block")
	}
	refusal := wantNotDefending(t, g.DeclareBlocker(oldGuard, attacker), oldGuard)
	if refusal.Defender != g.Seats[2].ID {
		t.Errorf("refusal's defender = %v, want seat 2 (the reselected one)", refusal.Defender)
	}
	if err := g.DeclareBlocker(newGuard, attacker); err != nil {
		t.Fatalf("the reselected walker's controller blocks: %v", err)
	}
}

// The record is combat-scoped state: an undo and a persisted restore
// carry it (so the block is still legal on either), and combat ending
// clears it.
func TestRecordedDefenderRidesCloneAndClearsWithCombat(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Doomed Walker", 4)
	guard := pushCombatant(t, g, g.Seats[1], "Guard", 1, 1)
	declaredAttackerAt(t, g, attacker, walker)
	removeFromBattlefieldForTest(t, g, walker, false)
	advanceTo(t, g, StepDeclareBlockers)

	if err := g.Clone().DeclareBlocker(guard, attacker); err != nil {
		t.Errorf("the clone forgot the recorded defender: %v", err)
	}
	// And a deploy restore mid-combat.
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
	if err := restored.DeclareBlocker(guard, attacker); err != nil {
		t.Errorf("the restored game forgot the recorded defender: %v", err)
	}
	passUntilStep(t, g, StepPostcombatMain)
	if g.attackDefenders != nil {
		t.Errorf("combat ended and the recorded defenders survived: %v", g.attackDefenders)
	}
}

// A recorded player who has left the game defends nothing (CR 800.4a).
func TestRecordedDefenderWhoLeftTheGameDefendsNothing(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Doomed Walker", 4)
	declaredAttackerAt(t, g, attacker, walker)
	removeFromBattlefieldForTest(t, g, walker, false)

	var before, after uuid.UUID
	g.WithWriteLock(func() {
		before = g.defendingPlayerForAttackerLocked(findBattlefieldCard(g, attacker))
		g.Seats[1].Eliminated = true
		after = g.defendingPlayerForAttackerLocked(findBattlefieldCard(g, attacker))
	})
	if before != g.Seats[1].ID {
		t.Fatalf("setup: recorded defender = %v, want seat 1", before)
	}
	if after != uuid.Nil {
		t.Errorf("an eliminated seat still defends: %v", after)
	}
}
