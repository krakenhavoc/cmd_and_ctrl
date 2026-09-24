package game

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// block_defender_test.go — #1339, CR 802.4a / 509.1a: "A defending
// player can block only with creatures they control. Those creatures
// can block only creatures attacking that player, a planeswalker that
// player controls, or a battle that player protects."
//
// Commander plays the attack multiple players option (CR 903.2), so
// there are up to three defending players in one combat, and a
// creature attacking one of them is not the others' to block. The
// option generator always knew that; the verb did not.

// fourSeatBlockTable is a four-seat table parked in declare_blockers:
// seat 0 attacks seat 1, seat 2 and seat 3 with one 3/3 each, and each
// defending seat has one untapped 1/4. attackers[i] attacks seats[i+1]
// and blockers[i] belongs to seats[i+1].
func fourSeatBlockTable(t *testing.T) (g *Game, attackers, blockers [3]uuid.UUID) {
	t.Helper()
	g = newActiveGameWithSeats(t, 4)
	for i := range 3 {
		attackers[i] = pushCombatant(t, g, g.Seats[0], "Attacker "+string(rune('A'+i)), 3, 3)
		blockers[i] = pushCombatant(t, g, g.Seats[i+1], "Blocker "+string(rune('A'+i)), 1, 4)
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	for i, a := range attackers {
		if err := g.DeclareAttacker(a, g.Seats[i+1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	advanceIntoStep(t, g, StepDeclareBlockers)
	return g, attackers, blockers
}

// wantNotDefending asserts err is a not_defending refusal naming the
// blocker, and returns it.
func wantNotDefending(t *testing.T, err error, blocker uuid.UUID) *BlockRefusedError {
	t.Helper()
	var refusal *BlockRefusedError
	if !errors.As(err, &refusal) || refusal.Reason != BlockReasonNotDefending {
		t.Fatalf("err = %v, want a not_defending refusal", err)
	}
	if !errors.Is(err, ErrIllegalBlock) {
		t.Errorf("a not_defending refusal does not wrap ErrIllegalBlock")
	}
	if refusal.Blocker != blocker {
		t.Errorf("refusal names blocker %v, want %v", refusal.Blocker, blocker)
	}
	return refusal
}

// The issue's repro: seat 1's creature blocks the 3/3 attacking seat 2.
// Refused, nothing stored, and seat 2 takes the 3.
func TestBlockRefusesACreatureAttackingAnotherPlayer(t *testing.T) {
	g, attackers, blockers := fourSeatBlockTable(t)

	err := g.DeclareBlockers([]BlockDeclaration{{Blocker: blockers[0], Attacker: attackers[1]}})
	refusal := wantNotDefending(t, err, blockers[0])
	if refusal.Defender != g.Seats[2].ID {
		t.Errorf("refusal's defender = %v, want seat 2", refusal.Defender)
	}
	want := "Attacker B is attacking P3, so only P3 can block it."
	if got := refusal.Sentence(g.Seats[1].ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if got := refusal.Sentence(g.Seats[2].ID); got != "Attacker B is attacking you, so only you can block it." {
		t.Errorf("the defender's own sentence = %q", got)
	}
	if c := findCard(g, blockers[0]); c.BlockingTarget != uuid.Nil {
		t.Fatalf("a refused block was stored: blocking %v", c.BlockingTarget)
	}

	life2 := g.Seats[2].Life
	passUntilStep(t, g, StepCombatDamage)
	if g.Seats[2].Life != life2-3 {
		t.Errorf("seat 2's life %d -> %d; the unblockable-by-seat-1 3/3 should hit for 3", life2, g.Seats[2].Life)
	}
}

// Every (defending creature, attacker) pair at a four-seat table: the
// verb accepts exactly the pairs where the blocker's seat is the one
// being attacked, and the option generator offers exactly the pairs
// the verb accepts — both directions, so neither can drift.
func TestBlockVerbAndGeneratorAgreeAtAFourSeatTable(t *testing.T) {
	g, attackers, blockers := fourSeatBlockTable(t)

	offered := map[[2]uuid.UUID]bool{}
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, seat := range g.Seats {
			for _, opt := range g.BlockOptionsLocked(seat.ID, 4) {
				for _, d := range opt.Blocks {
					offered[[2]uuid.UUID{d.Blocker, d.Attacker}] = true
				}
			}
		}
	})
	for bi, b := range blockers {
		for ai, a := range attackers {
			legal := bi == ai
			clone := g.Clone()
			err := clone.DeclareBlocker(b, a)
			if legal && err != nil {
				t.Errorf("seat %d blocking the attacker aimed at it: %v", bi+1, err)
			}
			if !legal {
				wantNotDefending(t, err, b)
			}
			if offered[[2]uuid.UUID{b, a}] != legal {
				t.Errorf("blocker of seat %d on attacker aimed at seat %d: offered = %v, verb legal = %v",
					bi+1, ai+1, offered[[2]uuid.UUID{b, a}], legal)
			}
		}
	}
}

// ALL OR NOTHING: one illegal entry refuses the whole set, legal half
// included (Decision 13).
func TestBlockSetWithOneWrongDefenderStoresNothing(t *testing.T) {
	g, attackers, blockers := fourSeatBlockTable(t)
	extra := pushCombatant(t, g, g.Seats[1], "Second Blocker", 1, 1)

	err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: blockers[0], Attacker: attackers[0]},
		{Blocker: extra, Attacker: attackers[2]},
	})
	wantNotDefending(t, err, extra)
	for _, id := range []uuid.UUID{blockers[0], extra} {
		if c := findCard(g, id); c.BlockingTarget != uuid.Nil {
			t.Errorf("%s was stored from a refused set", c.Name)
		}
	}
}

// A planeswalker attack is defended by the walker's CONTROLLER.
func TestBlockOnAPlaneswalkerAttackBelongsToItsController(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[3].ID, "Test Walker", 4)
	controllers := pushCombatant(t, g, g.Seats[3], "Walker's Guard", 1, 4)
	bystander := pushCombatant(t, g, g.Seats[1], "Bystander", 1, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker at the walker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	refusal := wantNotDefending(t, g.DeclareBlocker(bystander, attacker), bystander)
	want := "Attacker is attacking Test Walker, a planeswalker P4 controls, so only P4 can block it."
	if got := refusal.Sentence(g.Seats[1].ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareBlocker(controllers, attacker); err != nil {
		t.Fatalf("the walker's controller blocks: %v", err)
	}
}

// A battle attack is defended by its PROTECTOR, not its controller
// (CR 310.9d) — the one case where "the player whose permanent it is"
// is the wrong answer.
func TestBlockOnABattleAttackBelongsToItsProtector(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	battle := pushBattleForTest(g, g.Seats[1].ID, g.Seats[2].ID, "Test Siege", 5)
	controllers := pushCombatant(t, g, g.Seats[1], "Battle's Controller", 1, 4)
	protectors := pushCombatant(t, g, g.Seats[2], "Battle's Protector", 1, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, battle); err != nil {
		t.Fatalf("DeclareAttacker at the battle: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	refusal := wantNotDefending(t, g.DeclareBlocker(controllers, attacker), controllers)
	want := "Attacker is attacking Test Siege, a battle P3 protects, so only P3 can block it."
	if got := refusal.Sentence(g.Seats[1].ID); got != want {
		t.Errorf("sentence = %q, want %q", got, want)
	}
	if err := g.DeclareBlocker(protectors, attacker); err != nil {
		t.Fatalf("the battle's protector blocks: %v", err)
	}
}

// A creature attacking nothing has no defending player. Two shapes:
// one that is not attacking at all (the sandbox's old "pre-emptive"
// block) and one whose planeswalker left the battlefield (CR 506.4c).
// The generator offers neither; the verb now refuses both.
func TestBlockOnACreatureAttackingNothingIsRefused(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	idle := pushCombatant(t, g, g.Seats[0], "Idle", 2, 2)
	attacker := pushCombatant(t, g, g.Seats[0], "Attacker", 3, 3)
	walker := pushPlaneswalkerForTest(g, g.Seats[1].ID, "Doomed Walker", 1)
	blocker := pushCombatant(t, g, g.Seats[1], "Blocker", 1, 4)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if err := g.DeclareAttacker(attacker, walker); err != nil {
		t.Fatalf("DeclareAttacker: %v", err)
	}
	advanceIntoStep(t, g, StepDeclareBlockers)

	refusal := wantNotDefending(t, g.DeclareBlocker(blocker, idle), blocker)
	if got := refusal.Sentence(g.Seats[1].ID); got != "Idle isn't attacking anything, so no one can block it." {
		t.Errorf("sentence = %q", got)
	}

	g.WithWriteLock(func() {
		for i := range g.Battlefield.Cards {
			if g.Battlefield.Cards[i].InstanceID == walker {
				g.Battlefield.Cards = append(g.Battlefield.Cards[:i], g.Battlefield.Cards[i+1:]...)
				break
			}
		}
	})
	var offered int
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		offered = len(g.BlockOptionsLocked(g.Seats[1].ID, 4))
	})
	if offered != 0 {
		t.Errorf("the generator offers %d blocks on an attacker whose walker is gone", offered)
	}
	wantNotDefending(t, g.DeclareBlocker(blocker, attacker), blocker)
}

// #1343's reselect path, CR 508.7a + 509.1h. The attacker was blocked
// by seat 1, then reselected onto seat 2. The standing block stays and
// may be repeated (alone or beside a new pairing); seat 1 may not add
// a NEW blocker; seat 2, its defender now, may.
func TestBlockAfterAReselectFollowsTheNewDefender(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushCombatant(t, g, g.Seats[0], "Redirected", 3, 3)
	other := pushCombatant(t, g, g.Seats[0], "Still At Seat 1", 2, 2)
	oldBlocker := pushCombatant(t, g, g.Seats[1], "Old Defender's", 1, 4)
	oldSecond := pushCombatant(t, g, g.Seats[1], "Old Defender's Second", 1, 4)
	newBlocker := pushCombatant(t, g, g.Seats[2], "New Defender's", 1, 4)

	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, a := range []uuid.UUID{attacker, other} {
		if err := g.DeclareAttacker(a, g.Seats[1].ID); err != nil {
			t.Fatalf("DeclareAttacker: %v", err)
		}
	}
	g.WithWriteLock(func() { g.runStateChecksLocked() })
	advanceIntoStep(t, g, StepDeclareBlockers)
	if err := g.DeclareBlocker(oldBlocker, attacker); err != nil {
		t.Fatalf("the original defender blocks: %v", err)
	}
	lockInBlockDeclaration(t, g)
	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}

	// The standing block is not re-judged: repeating it is a no-op
	// success, alone and inside a set with a new legal pairing.
	if err := g.DeclareBlocker(oldBlocker, attacker); err != nil {
		t.Errorf("repeating the standing block: %v", err)
	}
	if err := g.DeclareBlockers([]BlockDeclaration{
		{Blocker: oldBlocker, Attacker: attacker},
		{Blocker: oldSecond, Attacker: other},
	}); err != nil {
		t.Errorf("a set repeating the standing block beside a legal new one: %v", err)
	}
	if c := findCard(g, oldBlocker); c.BlockingTarget != attacker {
		t.Errorf("the standing block moved: blocking %v", c.BlockingTarget)
	}

	// A NEW pairing from the old defender is refused.
	fresh := pushCombatant(t, g, g.Seats[1], "Late Arrival", 1, 4)
	refusal := wantNotDefending(t, g.DeclareBlocker(fresh, attacker), fresh)
	if refusal.Defender != g.Seats[2].ID {
		t.Errorf("refusal's defender = %v, want seat 2 (the reselected one)", refusal.Defender)
	}

	// The new defender may block it, and the generator agrees.
	var newOffered, oldOffered bool
	g.WithWriteLock(func() {
		g.RecomputeLayersIfStaleLocked()
		for _, opt := range g.BlockOptionsLocked(g.Seats[2].ID, 4) {
			for _, d := range opt.Blocks {
				newOffered = newOffered || (d.Blocker == newBlocker && d.Attacker == attacker)
			}
		}
		for _, opt := range g.BlockOptionsLocked(g.Seats[1].ID, 4) {
			for _, d := range opt.Blocks {
				oldOffered = oldOffered || d.Attacker == attacker
			}
		}
	})
	if !newOffered || oldOffered {
		t.Errorf("generator: new defender offered = %v (want true), old defender offered = %v (want false)", newOffered, oldOffered)
	}
	if err := g.DeclareBlocker(newBlocker, attacker); err != nil {
		t.Errorf("the new defender blocks the reselected attacker: %v", err)
	}
}

// Reselected BEFORE any block: the old defender is refused outright
// and the new one accepted — the verb half of
// TestReselectedAttackerIsBlockedAndDealsDamageAsTheNewTarget, which
// asserts only the generator.
func TestBlockAfterAReselectBeforeBlocksRefusesTheOldDefender(t *testing.T) {
	g := newActiveGameWithSeats(t, 4)
	attacker := pushReadyAttackerForTest(g, g.Seats[0].ID, "Attacker", 3)
	declaredAttackerAt(t, g, attacker, g.Seats[1].ID)
	if err := reselect(g, attacker, g.Seats[2].ID); err != nil {
		t.Fatalf("ReselectAttackTargetForEffect: %v", err)
	}
	oldBlocker := pushCombatant(t, g, g.Seats[1], "Old Defender's", 1, 1)
	newBlocker := pushCombatant(t, g, g.Seats[2], "New Defender's", 1, 1)
	advanceTo(t, g, StepDeclareBlockers)

	wantNotDefending(t, g.DeclareBlocker(oldBlocker, attacker), oldBlocker)
	if err := g.DeclareBlocker(newBlocker, attacker); err != nil {
		t.Errorf("the new defender blocks: %v", err)
	}
}
