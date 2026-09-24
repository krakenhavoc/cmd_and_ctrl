package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// lethal_matching_test.go pins #1261's policy half: the all-in lethal
// check counts only the blocks the defender can actually make.
//
// The nightly's heuristic gate stopped a game at turn 51 with two
// seats alive behind full, untapped boards — 22 creatures a side, one
// side holding 7 Drakes (3/3 flying) against the other's 5, the other
// at 5 life. Swinging everything was lethal: only a flier can block a
// flier, so two Drakes connect whatever the defender does. lethalPush
// said no, because it treated the defender's untapped creatures as a
// pool any attacker could be blocked out of — 22 attackers against 22
// "blockers", nothing through — and the table sat there until the turn
// budget ran out. A Bear cannot block a Drake, and the count has to
// know that.

// TestLethalPushCountsOnlyTheBlocksThatExist is that board at a size a
// reader can check by hand: three Drakes and three Ogres against two
// Drakes and four Ogres, the defender at 3. Six attackers, six
// untapped defenders — the pooled count says every attacker is
// blocked — but only two of the defenders can block a Drake, so one
// Drake connects for 3.
func TestLethalPushCountsOnlyTheBlocksThatExist(t *testing.T) {
	flier := keywords("flying")
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(6)), newSeat(1, withLife(3))},
		withBattlefield(
			creature(cardID(10), 0, "Drake", 3, 3, flier),
			creature(cardID(11), 0, "Drake", 3, 3, flier),
			creature(cardID(12), 0, "Drake", 3, 3, flier),
			creature(cardID(13), 0, "Ogre", 4, 4),
			creature(cardID(14), 0, "Ogre", 4, 4),
			creature(cardID(15), 0, "Ogre", 4, 4),
			creature(cardID(20), 1, "Drake", 3, 3, flier),
			creature(cardID(21), 1, "Drake", 3, 3, flier),
			creature(cardID(22), 1, "Ogre", 4, 4),
			creature(cardID(23), 1, "Ogre", 4, 4),
			creature(cardID(24), 1, "Ogre", 4, 4),
			creature(cardID(25), 1, "Ogre", 4, 4),
		),
		withTurn(20, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0),
		attackMove(t, 0, cardID(10), 1), attackMove(t, 0, cardID(11), 1), attackMove(t, 0, cardID(12), 1),
		attackMove(t, 0, cardID(13), 1), attackMove(t, 0, cardID(14), 1), attackMove(t, 0, cardID(15), 1),
	)
	d := decide(t, heuristic.New(), in)
	if got := chose(t, in, d); got == "Pass priority" {
		t.Fatalf("passed with a lethal swing on the board (%s); three Drakes against two flying blockers put 3 into a player at 3", d.Reason)
	}
	if d.Reason == "" || !strings.Contains(d.Reason, "all-in for the kill") {
		t.Errorf("attacked for %q; want the all-in lethal push", d.Reason)
	}
}

// TestLethalPushStillRespectsARealWall is the control: the same board
// with a third flying blocker. Now every Drake can be met, nothing is
// guaranteed through, and the push must not fire — a lethal check
// that over-counts is an alpha strike into the crack-back.
func TestLethalPushStillRespectsARealWall(t *testing.T) {
	flier := keywords("flying")
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(6)), newSeat(1, withLife(3))},
		withBattlefield(
			creature(cardID(10), 0, "Drake", 3, 3, flier),
			creature(cardID(11), 0, "Drake", 3, 3, flier),
			creature(cardID(12), 0, "Drake", 3, 3, flier),
			creature(cardID(13), 0, "Ogre", 4, 4),
			creature(cardID(14), 0, "Ogre", 4, 4),
			creature(cardID(15), 0, "Ogre", 4, 4),
			creature(cardID(20), 1, "Drake", 3, 3, flier),
			creature(cardID(21), 1, "Drake", 3, 3, flier),
			creature(cardID(22), 1, "Drake", 3, 3, flier),
			creature(cardID(23), 1, "Ogre", 4, 4),
			creature(cardID(24), 1, "Ogre", 4, 4),
			creature(cardID(25), 1, "Ogre", 4, 4),
		),
		withTurn(20, 0, "declare_attackers"),
	)
	in := input(0, v, passMove(0),
		attackMove(t, 0, cardID(10), 1), attackMove(t, 0, cardID(11), 1), attackMove(t, 0, cardID(12), 1),
		attackMove(t, 0, cardID(13), 1), attackMove(t, 0, cardID(14), 1), attackMove(t, 0, cardID(15), 1),
	)
	d := decide(t, heuristic.New(), in)
	if strings.Contains(d.Reason, "all-in for the kill") {
		t.Fatalf("pushed all-in (%q) into a board that can block every attacker", d.Reason)
	}
}
