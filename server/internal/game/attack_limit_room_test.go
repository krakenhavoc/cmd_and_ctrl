package game

import (
	"testing"

	"github.com/google/uuid"
)

// attack_limit_room_test.go — #1533, ADR 0045 Decision 46: the room a
// CR 508.1c count limit leaves, which the view publishes so the
// client's "attack with all" picker can cap its selection. The number
// is only worth publishing if it is EXACTLY the verb's answer: k new
// attackers at a target are accepted when k <= room and refused at
// room+1. These tests pin that agreement against the bulk verb, the
// one the picker sends.

// roomAt is AttackLimitRoomForEffect with a failure when no limit
// counts the target.
func roomAt(t *testing.T, g *Game, target uuid.UUID) int {
	t.Helper()
	room, limited := g.AttackLimitRoomForEffect(target)
	if !limited {
		t.Fatalf("no limit counts an attack on %s; want one", target)
	}
	return room
}

// bulkAt declares `ids` at `target` through the bulk verb on a CLONE,
// so the caller can probe k and k+1 against the same board.
func bulkAt(g *Game, ids []uuid.UUID, target uuid.UUID) error {
	decls := make([]AttackDeclaration, len(ids))
	for i, id := range ids {
		decls[i] = AttackDeclaration{Attacker: id, Target: target}
	}
	_, err := g.Clone().DeclareAttackers(decls)
	return err
}

// assertRoomIsTheVerbsAnswer checks room against the bulk verb:
// `room` new creatures at `target` are accepted, `room+1` refused.
func assertRoomIsTheVerbsAnswer(t *testing.T, g *Game, idle []uuid.UUID, target uuid.UUID) {
	t.Helper()
	room := roomAt(t, g, target)
	if room+1 > len(idle) {
		t.Fatalf("room %d leaves too few idle creatures (%d) to probe room+1", room, len(idle))
	}
	if room > 0 {
		if err := bulkAt(g, idle[:room], target); err != nil {
			t.Errorf("room is %d but %d attackers were refused: %v", room, room, err)
		}
	}
	attackLimitErr(t, bulkAt(g, idle[:room+1], target))
}

// TestAttackLimitRoomEachCombat — Silent Arbiter. Room 1 before anyone
// attacks, 0 after one creature does, and the same answer at every
// target, because an each-combat limit counts every attack.
func TestAttackLimitRoomEachCombat(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newFourPlayerActiveGame(t)
	me := g.Seats[0]
	pushLimitSource(t, g, g.Seats[1], "Silent Arbiter")
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, pushCombatant(t, g, me, "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	for _, seat := range g.Seats[1:] {
		if got := roomAt(t, g, seat.ID); got != 1 {
			t.Errorf("room at %s = %d, want 1", seat.Name, got)
		}
	}
	assertRoomIsTheVerbsAnswer(t, g, bears, g.Seats[2].ID)

	if err := g.DeclareAttacker(bears[0], g.Seats[1].ID); err != nil {
		t.Fatalf("the first attacker is legal: %v", err)
	}
	for _, seat := range g.Seats[1:] {
		if got := roomAt(t, g, seat.ID); got != 0 {
			t.Errorf("room at %s after one attack = %d, want 0", seat.Name, got)
		}
	}
	assertRoomIsTheVerbsAnswer(t, g, bears[1:], g.Seats[3].ID)
}

// TestAttackLimitRoomAttackingYou — Crawlspace in a four-seat game. The
// room is the protected seat's alone: the other opponents are not
// limited at all, and attacks on them do not spend it. A planeswalker
// the protected seat controls is not "you".
func TestAttackLimitRoomAttackingYou(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitAttackingYou, Max: 2}}, nil)
	g := newFourPlayerActiveGame(t)
	me, crawl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushLimitSource(t, g, crawl, "Crawlspace")
	walker := pushPlaneswalkerForTest(g, crawl.ID, "Their Walker", 4)
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, pushCombatant(t, g, me, "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	if _, limited := g.AttackLimitRoomForEffect(other.ID); limited {
		t.Error("an opponent without Crawlspace reads as limited")
	}
	if _, limited := g.AttackLimitRoomForEffect(walker); limited {
		t.Error("the Crawlspace player's planeswalker reads as limited")
	}
	if got := roomAt(t, g, crawl.ID); got != 2 {
		t.Errorf("room at the Crawlspace player = %d, want 2", got)
	}

	if err := g.DeclareAttacker(bears[0], other.ID); err != nil {
		t.Fatalf("attack on the other opponent: %v", err)
	}
	if got := roomAt(t, g, crawl.ID); got != 2 {
		t.Errorf("an attack elsewhere spent the Crawlspace player's room: %d", got)
	}
	if err := g.DeclareAttacker(bears[1], crawl.ID); err != nil {
		t.Fatalf("attack on the Crawlspace player: %v", err)
	}
	if got := roomAt(t, g, crawl.ID); got != 1 {
		t.Errorf("room after one attack on the Crawlspace player = %d, want 1", got)
	}
	assertRoomIsTheVerbsAnswer(t, g, bears[2:], crawl.ID)
}

// TestAttackLimitRoomTheTightestWins — Silent Arbiter and Crawlspace
// together: the Crawlspace player's room is the smaller allowance.
func TestAttackLimitRoomTheTightestWins(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{
		{Scope: AttackLimitAttackingYou, Max: 2},
		{Scope: AttackLimitEachCombat, Max: 3},
	}, nil)
	g := newFourPlayerActiveGame(t)
	me, crawl, other := g.Seats[0], g.Seats[1], g.Seats[2]
	pushLimitSource(t, g, crawl, "Both")
	var bears []uuid.UUID
	for i := 0; i < 5; i++ {
		bears = append(bears, pushCombatant(t, g, me, "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)

	if got := roomAt(t, g, crawl.ID); got != 2 {
		t.Errorf("room at the protected seat = %d, want 2 (the you-limit)", got)
	}
	if got := roomAt(t, g, other.ID); got != 3 {
		t.Errorf("room elsewhere = %d, want 3 (the each-combat limit)", got)
	}
	// Two attacks elsewhere leave the each-combat limit one slot, which
	// is now tighter than the you-limit's two.
	for _, b := range bears[:2] {
		if err := g.DeclareAttacker(b, other.ID); err != nil {
			t.Fatalf("attack elsewhere: %v", err)
		}
	}
	if got := roomAt(t, g, crawl.ID); got != 1 {
		t.Errorf("room at the protected seat = %d, want 1 (the each-combat limit)", got)
	}
	assertRoomIsTheVerbsAnswer(t, g, bears[2:], crawl.ID)
}

// TestAttackLimitRoomNeverNegative — a limit that arrives after the
// attacks were declared (Decision 44: it unmakes nothing) is over its
// Max, and its room is 0, not negative: every new attacker raises the
// count and is refused.
func TestAttackLimitRoomNeverNegative(t *testing.T) {
	stubCombatLimits(t, []AttackLimit{{Scope: AttackLimitEachCombat, Max: 1}}, nil)
	g := newActiveGame(t)
	var bears []uuid.UUID
	for i := 0; i < 4; i++ {
		bears = append(bears, pushCombatant(t, g, g.Seats[0], "Bear", 2, 2))
	}
	advanceIntoStep(t, g, StepDeclareAttackers)
	for _, b := range bears[:2] {
		if err := g.DeclareAttacker(b, g.Seats[1].ID); err != nil {
			t.Fatalf("attack before the limit: %v", err)
		}
	}
	pushLimitSource(t, g, g.Seats[1], "Late Arbiter")

	if got := roomAt(t, g, g.Seats[1].ID); got != 0 {
		t.Errorf("room under an over-full limit = %d, want 0", got)
	}
	assertRoomIsTheVerbsAnswer(t, g, bears[2:], g.Seats[1].ID)
}

// TestAttackLimitRoomUnlimited — no limit on the battlefield: nothing
// reads as limited, which is what keeps the wire field absent at
// nearly every table.
func TestAttackLimitRoomUnlimited(t *testing.T) {
	stubCombatLimits(t, nil, nil)
	g := newActiveGame(t)
	pushCombatant(t, g, g.Seats[0], "Bear", 2, 2)
	advanceIntoStep(t, g, StepDeclareAttackers)
	if _, limited := g.AttackLimitRoomForEffect(g.Seats[1].ID); limited {
		t.Error("a table with no attack limit reads as limited")
	}
}
