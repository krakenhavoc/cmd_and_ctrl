package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// blockset_decide_test.go pins #3066 at the decision: the defender's
// blocks are checked as a whole set against lethal damage
// (Config.BlockSetSurvives).

const setSurvives = "a set that survives"

// blockBoard: seat 1 attacks seat 0 with `atks`; seat 0 has `mine` to
// block with. Every pair is on offer, in the order given (attacker by
// attacker for each blocker), then "no blocks".
func blockBoard(t *testing.T, me protocol.PlayerView, mine, atks []protocol.CardView) aiseat.Input {
	t.Helper()
	cards := append(append([]protocol.CardView(nil), atks...), mine...)
	v := newView(
		[]protocol.PlayerView{me, newSeat(1)},
		withBattlefield(cards...),
		withTurn(9, 1, "declare_blockers"),
	)
	var moves []legal.Move
	for _, b := range mine {
		for _, a := range atks {
			moves = append(moves, blockMove(t, 0, b.InstanceID, a.InstanceID))
		}
	}
	return input(0, v, append(moves, finishBlocksMove(0))...)
}

// blockedWith decides the board with the switch as given, and reports
// which attacker the bot blocked ("" for none) and why.
func blockedWith(t *testing.T, in aiseat.Input, on bool) (string, string) {
	t.Helper()
	d := decideWith(t, in, func(c *heuristic.Config) { c.BlockSetSurvives = on })
	if d.Index < 0 || in.Moves[d.Index].Kind != legal.KindBlock {
		return "", d.Reason
	}
	return string(in.Moves[d.Index].Params), d.Reason
}

// TestBlockSetSurvivesTheCommanderSwing is the issue's board, from the
// full seed-1409 Wurm mirror: the defender on 3, its 3/3 commander its
// only creature, facing a Bear 2/2 and a 3/3 commander. Scored on its
// own the Bear block is better (+6.90: it kills the Bear and loses
// nothing), and the commander's 3 kill the defender. Commander on
// commander takes 2 and lives.
func TestBlockSetSurvivesTheCommanderSwing(t *testing.T) {
	bear := creature(cardID(20), 1, "Bear", 2, 2, attacking(0))
	theirs := creature(cardID(21), 1, "Chief", 3, 3, commander(), attacking(0))
	mine := creature(cardID(1), 0, "Chief", 3, 3, commander())
	in := blockBoard(t, newSeat(0, withLife(3)),
		[]protocol.CardView{mine}, []protocol.CardView{bear, theirs})

	got, reason := blockedWith(t, in, true)
	if !strings.Contains(got, theirs.InstanceID) || !strings.Contains(reason, setSurvives) {
		t.Fatalf("blocked %s (%s); want commander on commander, which takes 2 of 3", got, reason)
	}
	got, reason = blockedWith(t, in, false)
	if !strings.Contains(got, bear.InstanceID) || !strings.Contains(reason, "+6.90") {
		t.Fatalf("switch off: blocked %s (%s); want the old Bear block at +6.90", got, reason)
	}
}

// TestBlockSetSurvivesCommanderDamage: on 40 life, but their 2/2
// commander has dealt 19 already. A 3/3 Bear is the bigger kill, and
// the commander's 2 more are 21 (CR 903.10a).
func TestBlockSetSurvivesCommanderDamage(t *testing.T) {
	bear := creature(cardID(20), 1, "Bear", 3, 3, attacking(0))
	theirs := creature(cardID(21), 1, "Chief", 2, 2, commander(), attacking(0))
	me := newSeat(0, withLife(40))
	me.CommanderDamage = map[string]int{theirs.InstanceID: 19}
	in := blockBoard(t, me,
		[]protocol.CardView{creature(cardID(1), 0, "Ogre", 4, 4)}, []protocol.CardView{bear, theirs})

	if got, reason := blockedWith(t, in, true); !strings.Contains(got, theirs.InstanceID) || !strings.Contains(reason, setSurvives) {
		t.Fatalf("blocked %s (%s); the commander's 2 are its 21st", got, reason)
	}
	if got, reason := blockedWith(t, in, false); !strings.Contains(got, bear.InstanceID) {
		t.Fatalf("switch off: blocked %s (%s); want the bigger kill", got, reason)
	}
	// One point short of lethal, the bigger kill stands.
	me.CommanderDamage[theirs.InstanceID] = 18
	in = blockBoard(t, me,
		[]protocol.CardView{creature(cardID(1), 0, "Ogre", 4, 4)}, []protocol.CardView{bear, theirs})
	if got, reason := blockedWith(t, in, true); !strings.Contains(got, bear.InstanceID) {
		t.Fatalf("on 18 commander damage: blocked %s (%s); want the bigger kill", got, reason)
	}
}

// TestBlockSetSurvivesPoison: on 40 life with 8 poison counters,
// facing a Bear and a toxic 2 Rat, both 2/2. The two blocks score the
// same, and the Rat's 2 poison are the tenth counter (CR 104.3d).
func TestBlockSetSurvivesPoison(t *testing.T) {
	bear := creature(cardID(20), 1, "Bear", 2, 2, attacking(0))
	rat := creature(cardID(21), 1, "Rat", 2, 2, keywords("toxic 2"), attacking(0))
	me := newSeat(0, withLife(40))
	me.Poison = 8
	in := blockBoard(t, me,
		[]protocol.CardView{creature(cardID(1), 0, "Ogre", 4, 4)}, []protocol.CardView{bear, rat})

	if got, reason := blockedWith(t, in, true); !strings.Contains(got, rat.InstanceID) || !strings.Contains(reason, setSurvives) {
		t.Fatalf("blocked %s (%s); the Rat's toxic is the tenth poison counter", got, reason)
	}
	if got, reason := blockedWith(t, in, false); !strings.Contains(got, bear.InstanceID) {
		t.Fatalf("switch off: blocked %s (%s); want the first of two equal blocks", got, reason)
	}
}

// TestBlockSetSurvivesTrampleOverflow: on 5, facing a 6/6 trampling
// Wurm and a Bear, with one Ogre. The Ogre on the Bear lets the Wurm's
// 6 through; on the Wurm it lets 2 over and the Bear's 2, which is 4.
func TestBlockSetSurvivesTrampleOverflow(t *testing.T) {
	bear := creature(cardID(20), 1, "Bear", 2, 2, attacking(0))
	wurm := creature(cardID(21), 1, "Wurm", 6, 6, keywords("trample"), attacking(0))
	in := blockBoard(t, newSeat(0, withLife(5)),
		[]protocol.CardView{creature(cardID(1), 0, "Ogre", 4, 4)}, []protocol.CardView{bear, wurm})

	if got, reason := blockedWith(t, in, true); !strings.Contains(got, wurm.InstanceID) {
		t.Fatalf("blocked %s (%s); only the Ogre on the Wurm leaves the bot alive", got, reason)
	}
}

// TestBlockSetSurvivesWithTwoBlockers: the check looks past the block
// being declared. On 3, facing a Bear 2/2, an Ogre 4/4 and a 3/3
// commander, with a Bear and a Wall 0/5 that can block only the Ogre.
// The Bear has to take the commander and the Wall the Ogre, taking 2.
// With the Wall on the Ogre, my Bear on theirs lets the commander's 3
// through, and my Bear on the Ogre lets 5 through: both kill.
func TestBlockSetSurvivesWithTwoBlockers(t *testing.T) {
	bear := creature(cardID(20), 1, "Bear", 2, 2, attacking(0))
	ogre := creature(cardID(21), 1, "Ogre", 4, 4, attacking(0))
	chief := creature(cardID(22), 1, "Chief", 3, 3, commander(), attacking(0))
	myBear := creature(cardID(1), 0, "Bear", 2, 2)
	wall := creature(cardID(2), 0, "Wall", 0, 5)
	cards := []protocol.CardView{bear, ogre, chief, myBear, wall}
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(3)), newSeat(1)},
		withBattlefield(cards...),
		withTurn(9, 1, "declare_blockers"),
	)
	in := input(0, v,
		blockMove(t, 0, myBear.InstanceID, bear.InstanceID),
		blockMove(t, 0, myBear.InstanceID, ogre.InstanceID),
		blockMove(t, 0, myBear.InstanceID, chief.InstanceID),
		blockMove(t, 0, wall.InstanceID, ogre.InstanceID),
		finishBlocksMove(0),
	)
	d := decideWith(t, in, nil)
	if d.Index != 2 && d.Index != 3 {
		t.Fatalf("decided %q (%s); want the Bear on the commander or the Wall on the Ogre", chose(t, in, d), d.Reason)
	}
	if d.Index == 3 {
		// The Wall went first; the Bear must still take the commander.
		cards[4].BlockingTarget = ogre.InstanceID
		v.Battlefield.Cards = cards
		in = input(0, v,
			blockMove(t, 0, myBear.InstanceID, bear.InstanceID),
			blockMove(t, 0, myBear.InstanceID, ogre.InstanceID),
			blockMove(t, 0, myBear.InstanceID, chief.InstanceID),
			finishBlocksMove(0),
		)
		if d = decideWith(t, in, nil); d.Index != 2 {
			t.Fatalf("after the Wall: decided %q (%s); want the Bear on the commander", chose(t, in, d), d.Reason)
		}
	}
}

// TestBlockSetNoSurvivorKeepsTheOldBlock: on 3, facing two 3/3s with
// one blocker, every set dies, so the decision is the old one.
func TestBlockSetNoSurvivorKeepsTheOldBlock(t *testing.T) {
	in := blockBoard(t, newSeat(0, withLife(3)),
		[]protocol.CardView{creature(cardID(1), 0, "Chief", 3, 3, commander())},
		[]protocol.CardView{
			creature(cardID(20), 1, "Centaur", 3, 3, attacking(0)),
			creature(cardID(21), 1, "Chief", 3, 3, commander(), attacking(0)),
		})
	onGot, onReason := blockedWith(t, in, true)
	offGot, offReason := blockedWith(t, in, false)
	if onGot != offGot || onReason != offReason {
		t.Fatalf("on: %s (%s); off: %s (%s); want the same decision", onGot, onReason, offGot, offReason)
	}
}
