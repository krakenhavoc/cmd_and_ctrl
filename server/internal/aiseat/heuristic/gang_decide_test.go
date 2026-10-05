package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// gang_decide_test.go pins #1548 at the decision: the block planner
// scores a block onto an attacker that is already blocked as the whole
// group, and the attack planner looks past the two-turn race.

// gangView: seat 1 attacks seat 0 (at myLife) with atk; seat 0 already
// blocks it with `have` and has `spare` to add. The moves are the one
// block on offer and the declaring defender's "done".
func gangView(t *testing.T, myLife int, atk protocol.CardView, have []protocol.CardView, spare protocol.CardView) (protocol.GameView, string) {
	t.Helper()
	cards := []protocol.CardView{atk}
	for _, h := range have {
		h.BlockingTarget = atk.InstanceID
		cards = append(cards, h)
	}
	cards = append(cards, spare)
	return newView(
		[]protocol.PlayerView{newSeat(0, withLife(myLife)), newSeat(1)},
		withBattlefield(cards...),
		withTurn(5, 1, "declare_blockers"),
	), spare.InstanceID
}

// gangDecision decides the gang view and reports whether the bot added
// the spare, and why.
func gangDecision(t *testing.T, myLife int, atk protocol.CardView, have []protocol.CardView, spare protocol.CardView) (bool, string) {
	t.Helper()
	v, id := gangView(t, myLife, atk, have, spare)
	in := input(0, v, blockMove(t, 0, id, atk.InstanceID), finishBlocksMove(0))
	d := decide(t, heuristic.New(), in)
	return d.Index == 0, d.Reason
}

// TestGangBlockKillsTheChumpedWurm is the issue's board: one Ogre is
// already in front of a Wurm and a second is spare. 4+4 ≥ 7 kills the
// Wurm, and its 7 damage kills one Ogre either way, so the second Ogre
// costs nothing. Before #1548 the planner asked only whether the second
// Ogre killed the Wurm alone, and finished its declaration.
func TestGangBlockKillsTheChumpedWurm(t *testing.T) {
	added, reason := gangDecision(t, 40,
		creature(cardID(20), 1, "Wurm", 7, 7, keywords("trample"), attacking(0)),
		[]protocol.CardView{creature(cardID(10), 0, "Ogre", 4, 4)},
		creature(cardID(11), 0, "Ogre", 4, 4))
	if !added || !strings.Contains(reason, "gang block to kill") {
		t.Fatalf("added=%v (%s); a second Ogre on the chumped Wurm is a free kill", added, reason)
	}
}

// TestGangBlockFirstStrikerKillsFirst: a first striker deals its damage
// before the gang deals regular damage, and spends it on the blocker
// that would finish it. A 3/3 first striker blocked by a Bear kills a
// 3/3 that joins before it strikes, and the Bear's 2 do not kill it — so
// the 3/3 dies for nothing. An Ogre is too big to kill first, so the
// same join with an Ogre does kill it.
func TestGangBlockFirstStrikerKillsFirst(t *testing.T) {
	knight := creature(cardID(20), 1, "Knight", 3, 3, keywords("first strike"), attacking(0))
	bear := []protocol.CardView{creature(cardID(10), 0, "Bear", 2, 2)}
	if added, reason := gangDecision(t, 40, knight, bear, creature(cardID(11), 0, "Centaur", 3, 3)); added {
		t.Errorf("added the 3/3 (%s); the Knight kills it first and lives", reason)
	}
	if added, reason := gangDecision(t, 40, knight, bear, creature(cardID(11), 0, "Ogre", 4, 4)); !added || !strings.Contains(reason, "to kill") {
		t.Errorf("added=%v (%s); the Knight cannot kill an Ogre first, so Bear and Ogre kill it", added, reason)
	}
}

// TestGangBlockDeathtouch: a deathtoucher needs 1 point per blocker, so
// how many joiners it kills is its power. A 1/3 deathtoucher chumped by
// a Bear kills only one Bear, so a second Bear is a free kill; a 2/3
// kills an Ogre that joins as well, which costs more than the kill.
func TestGangBlockDeathtouch(t *testing.T) {
	bear := []protocol.CardView{creature(cardID(10), 0, "Bear", 2, 2)}
	if added, reason := gangDecision(t, 40,
		creature(cardID(20), 1, "Adder", 1, 3, keywords("deathtouch"), attacking(0)),
		bear, creature(cardID(11), 0, "Bear", 2, 2)); !added || !strings.Contains(reason, "to kill") {
		t.Errorf("added=%v (%s); the 1-power deathtoucher kills one of the two Bears and dies", added, reason)
	}
	if added, reason := gangDecision(t, 40,
		creature(cardID(20), 1, "Adder", 2, 3, keywords("deathtouch"), attacking(0)),
		bear, creature(cardID(11), 0, "Ogre", 4, 4)); added {
		t.Errorf("added the Ogre (%s); a 2-power deathtoucher kills it too", reason)
	}
}

// TestGangBlockSoaksTrampleOverflow: at 4 life a Wurm a Bear is chumping
// still tramples 5 over, which kills. A second Bear kills nothing — 4 <
// 7 — but soaks 2 more and leaves the bot on 1. Before #1548 the Wurm
// counted as stopped, so the bot was not even desperate, and a block on
// a blocked attacker was worth nothing unless it killed.
func TestGangBlockSoaksTrampleOverflow(t *testing.T) {
	added, reason := gangDecision(t, 4,
		creature(cardID(20), 1, "Wurm", 7, 7, keywords("trample"), attacking(0)),
		[]protocol.CardView{creature(cardID(10), 0, "Bear", 2, 2)},
		creature(cardID(11), 0, "Bear", 2, 2))
	if !added {
		t.Fatalf("finished (%s) with 5 trample damage about to kill the bot on 4", reason)
	}
	// Without trample the chump holds the attacker in full: nothing to
	// soak, nothing to kill.
	if added, reason := gangDecision(t, 4,
		creature(cardID(20), 1, "Giant", 7, 7, attacking(0)),
		[]protocol.CardView{creature(cardID(10), 0, "Bear", 2, 2)},
		creature(cardID(11), 0, "Bear", 2, 2)); added {
		t.Errorf("added a Bear to a chumped Giant (%s); it changes nothing", reason)
	}
}

// --- the attrition horizon ----------------------------------------

const attritionLabel = "attrition"

// wurmMirror is the full seed-1409 Wurm-edge board after the Drakes
// have traded off: the bot's two Wurms against one, six Ogres, five
// Bears and a 3/3 on each side, both players on 6 — with `extra` added
// to the defender.
func wurmMirror(myWurms int, extra ...protocol.CardView) (protocol.GameView, []string) {
	var cards []protocol.CardView
	n := 0
	add := func(seat int, name string, p, tough int, opts ...cardOpt) {
		n++
		cards = append(cards, creature(cardID(n), seat, name, p, tough, opts...))
	}
	for seat, wurms := range []int{myWurms, 1} {
		for i := 0; i < wurms; i++ {
			add(seat, "Wurm", 7, 7, keywords("trample"))
		}
		for i := 0; i < 6; i++ {
			add(seat, "Ogre", 4, 4)
		}
		for i := 0; i < 5; i++ {
			add(seat, "Bear", 2, 2)
		}
		add(seat, "Chief", 3, 3)
	}
	cards = append(cards, extra...)
	var mine []string
	for _, c := range cards {
		if c.Controller == seatID(0).String() {
			mine = append(mine, c.InstanceID)
		}
	}
	return newView(
		[]protocol.PlayerView{newSeat(0, withLife(6)), newSeat(1, withLife(6))},
		withBattlefield(cards...),
		withTurn(20, 0, "declare_attackers"),
	), mine
}

// TestAttritionSwingsWhenTheRaceCannot is the board that stalled once
// the race assumed gang blocks. There is no two-turn race: a gang-
// blocking defender holds every swing to 2 against 6 life. But one Wurm
// into theirs is an even trade that keeps the edge, and the board after
// it races. The bot sends exactly that Wurm and keeps everything else
// home.
func TestAttritionSwingsWhenTheRaceCannot(t *testing.T) {
	v, mine := wurmMirror(2)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if !strings.Contains(d.Reason, attritionLabel) {
		t.Fatalf("decided %q (%s); want an attrition swing", chose(t, in, d), d.Reason)
	}
	if got := chose(t, in, d); !strings.Contains(got, cardID(1)) {
		t.Errorf("attacked with %q; the line starts with a Wurm", got)
	}
	// Once the Wurm is declared, the plan is complete: everything else
	// is the reserve.
	v.Battlefield.Cards[0].AttackingTarget = seatID(1).String()
	v.Battlefield.Cards[0].Tapped = true
	in = input(0, v, attackAll(t, mine[1:])...)
	if d := decide(t, heuristic.New(), in); !strings.Contains(chose(t, in, d), "Pass") {
		t.Errorf("after the Wurm, decided %q (%s); the rest stays home", chose(t, in, d), d.Reason)
	}
}

// smallBoard is seat 0 (the bot, to attack, on myLife) against seat 1
// on theirLife, every creature untapped, and attack moves for all of
// the bot's.
func smallBoard(t *testing.T, myLife, theirLife int, cards ...protocol.CardView) (protocol.GameView, []legal.Move) {
	t.Helper()
	var mine []string
	for _, c := range cards {
		if c.Controller == seatID(0).String() {
			mine = append(mine, c.InstanceID)
		}
	}
	return newView(
		[]protocol.PlayerView{newSeat(0, withLife(myLife)), newSeat(1, withLife(theirLife))},
		withBattlefield(cards...),
		withTurn(20, 0, "declare_attackers"),
	), attackAll(t, mine)
}

// TestAttritionIgnoresDamageTheDefenderStopsForFree: a lone Ogre into a
// defender on 9 holding its commander and a first-strike Knight. The
// race's model has the defender take the 4, since it survives that. But
// the commander can chump the Ogre for nothing — it goes back to the
// command zone and is cast again — and then nothing has happened and
// the same swing comes round next turn. That loop is what ran a
// commander into a commander for a dozen turns at the end of the Wurm
// mirror, so damage the defender can stop for free is not progress.
func TestAttritionIgnoresDamageTheDefenderStopsForFree(t *testing.T) {
	v, moves := smallBoard(t, 5, 9,
		creature(cardID(1), 0, "Knight", 2, 2, keywords("first strike")),
		creature(cardID(2), 0, "Ogre", 4, 4),
		creature(cardID(3), 0, "Adder", 1, 1, keywords("deathtouch")),
		creature(cardID(4), 1, "Chief", 3, 3, commander()),
		creature(cardID(5), 1, "Knight", 2, 2, keywords("first strike")),
	)
	in := input(0, v, moves...)
	if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, attritionLabel) {
		t.Fatalf("decided %q (%s); their commander blocks the Ogre for free", chose(t, in, d), d.Reason)
	}
}

// TestAttritionWillNotTradeDown: the defender is on 2 behind a Drake, a
// first-strike Knight and a 1/1 deathtoucher; the bot has an Ogre, a
// Knight and a Giant. The Giant into the deathtoucher is one for one
// and the board after it can be raced, but it trades the bot's best
// creature for their worst and leaves it behind on the board. The edge
// has to survive every exchange.
func TestAttritionWillNotTradeDown(t *testing.T) {
	v, moves := smallBoard(t, 9, 2,
		creature(cardID(1), 0, "Ogre", 4, 4),
		creature(cardID(2), 0, "Knight", 2, 2, keywords("first strike")),
		creature(cardID(3), 0, "Giant", 5, 5),
		creature(cardID(4), 1, "Drake", 3, 3, keywords("flying")),
		creature(cardID(5), 1, "Adder", 1, 1, keywords("deathtouch")),
		creature(cardID(6), 1, "Knight", 2, 2, keywords("first strike")),
	)
	in := input(0, v, moves...)
	if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, attritionLabel) {
		t.Fatalf("decided %q (%s); the line trades the Giant for the deathtoucher", chose(t, in, d), d.Reason)
	}
}

// TestAttritionNeedsAnEdge: the perfect mirror has no attrition line —
// every trade is even and nothing is ever left over — and neither does
// the edge board once the defender has a deathtoucher too: the Wurm
// that was the edge would trade for a 1/1, and nothing after that
// races. Neither swings on the horizon.
func TestAttritionNeedsAnEdge(t *testing.T) {
	for _, tc := range []struct {
		name  string
		wurms int
		extra []protocol.CardView
	}{
		{"a perfect mirror", 1, nil},
		{"the edge matched by a deathtoucher", 2, []protocol.CardView{creature(cardID(90), 1, "Adder", 1, 1, keywords("deathtouch"))}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, mine := wurmMirror(tc.wurms, tc.extra...)
			in := input(0, v, attackAll(t, mine)...)
			if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, attritionLabel) {
				t.Fatalf("decided %q (%s); there is no edge to grind", chose(t, in, d), d.Reason)
			}
		})
	}
}
