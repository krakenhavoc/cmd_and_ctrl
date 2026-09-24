package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_test.go pins #1409's policy half: the two-turn race term in the
// attack decision (race.go).
//
// The board is the one the issue describes: two seats under 7 life
// behind full, untapped, EVEN boards — the evasion classes matched as
// well as the bodies — where no all-in is lethal this turn and every
// single attack is blocked at a loss. Before the race term the
// heuristic sat behind that board until a library ran out. With one
// Drake more on one side, that side has a win it can see two turns
// out; with none, it has nothing, and the bot must not invent one.

const raceLabel = "two-turn race"

// raceBoard builds seat 0 (the bot, to attack) and seat 1 at the given
// life totals: each side the given number of Drakes (3/3 flying), then
// five Wurms (7/7 trample) and four Ogres (4/4), in that order.
func raceBoard(myDrakes, theirDrakes, myLife, theirLife int, extra ...protocol.CardView) (protocol.GameView, []string) {
	flier := keywords("flying")
	trample := keywords("trample")
	var cards []protocol.CardView
	var mine []string
	id := 100
	add := func(ctrl int, name string, p, tough int, opts ...cardOpt) {
		c := creature(cardID(id), ctrl, name, p, tough, opts...)
		cards = append(cards, c)
		if ctrl == 0 {
			mine = append(mine, c.InstanceID)
		}
		id++
	}
	for ctrl, drakes := range []int{myDrakes, theirDrakes} {
		for i := 0; i < drakes; i++ {
			add(ctrl, "Drake", 3, 3, flier)
		}
		for i := 0; i < 5; i++ {
			add(ctrl, "Wurm", 7, 7, trample)
		}
		for i := 0; i < 4; i++ {
			add(ctrl, "Ogre", 4, 4)
		}
	}
	cards = append(cards, extra...)
	seats := []protocol.PlayerView{newSeat(0, withLife(myLife)), newSeat(1, withLife(theirLife))}
	for _, c := range extra {
		if c.Controller == seatID(2).String() {
			// Far out of reach, so the third seat is never a lethalPush
			// target of its own: all it contributes is its crack-back.
			seats = append(seats, newSeat(2, withLife(100)))
			break
		}
	}
	return newView(seats, withBattlefield(cards...), withTurn(30, 0, "declare_attackers")), mine
}

// attackAll is every creature of seat 0 offered against seat 1, plus
// the pass.
func attackAll(t *testing.T, ids []string) []legal.Move {
	moves := []legal.Move{passMove(0)}
	for _, id := range ids {
		moves = append(moves, attackMove(t, 0, id, 1))
	}
	return moves
}

// TestRaceSwingsTheDrakeThatIsLeftOver is the board the issue names,
// with the edge that makes it winnable: seven Drakes against six, the
// ground matched Wurm for Wurm and Ogre for Ogre, both players at 6.
// All-in puts one Drake through for 3, which is not lethal, so
// lethalPush stays quiet. But one Drake now is 3, they are at 3, and
// the same all-in next turn is lethal — while nothing they have can
// get past the creatures kept home. The bot has to see that and swing.
func TestRaceSwingsTheDrakeThatIsLeftOver(t *testing.T) {
	v, mine := raceBoard(7, 6, 6, 6)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if got := chose(t, in, d); got == "Pass priority" {
		t.Fatalf("passed on a board it wins in two turns (%s): a Drake up with the ground matched", d.Reason)
	}
	if !strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("attacked for %q; want the two-turn race", d.Reason)
	}
	if name := v.Battlefield.Cards[indexOfCard(v, in.Moves[d.Index])].Name; name != "Drake" {
		t.Errorf("the race sent a %s first; the evasive attacker is the one the plan is built on", name)
	}
}

// TestPerfectMirrorHasNoRace is the control: the same board with the
// Drakes even. Every trade is even and nothing is left over for next
// turn, so there is no race to find and the bot must not attack into
// it — a race term that fires on a mirror is an alpha strike into the
// crack-back.
func TestPerfectMirrorHasNoRace(t *testing.T) {
	v, mine := raceBoard(6, 6, 6, 6)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("raced a perfect mirror (%s)", d.Reason)
	}
}

// TestRaceNeverSwingsIntoALethalCrackBack: three Drakes against a
// defender at 10 with one flying blocker. Every Drake swinging, twice,
// kills — 9 now, 6 next turn — but three Ogres and a Drake come back
// at a player on 4 with nothing home to stop them. Swinging fewer
// keeps too little home for the same reason. The race must not fire.
func TestRaceNeverSwingsIntoALethalCrackBack(t *testing.T) {
	flier := keywords("flying")
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(4)), newSeat(1, withLife(10))},
		withBattlefield(
			creature(cardID(10), 0, "Drake", 3, 3, flier),
			creature(cardID(11), 0, "Drake", 3, 3, flier),
			creature(cardID(12), 0, "Drake", 3, 3, flier),
			creature(cardID(20), 1, "Ogre", 4, 4),
			creature(cardID(21), 1, "Ogre", 4, 4),
			creature(cardID(22), 1, "Ogre", 4, 4),
			creature(cardID(23), 1, "Drake", 3, 3, flier),
		),
		withTurn(20, 0, "declare_attackers"),
	)
	in := input(0, v, attackAll(t, []string{cardID(10), cardID(11), cardID(12)})...)
	d := decide(t, heuristic.New(), in)
	if strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("raced into a lethal crack-back (%s)", d.Reason)
	}
}

// TestRaceCountsEveryOpponentsCrackBack: the winning board from
// TestRaceSwingsTheDrakeThatIsLeftOver, with a third seat holding a
// 6/6. While the 6/6 can be blocked, what the bot keeps home holds it
// and the race is on. Make it unblockable and the third seat kills the
// bot from 6 on the turn in between, whatever it keeps home — so there
// is no race, however good it looks against seat 1 alone.
func TestRaceCountsEveryOpponentsCrackBack(t *testing.T) {
	rogue := creature(cardID(900), 2, "Rogue", 6, 6)
	v, mine := raceBoard(7, 6, 6, 6, rogue)
	in := input(0, v, attackAll(t, mine)...)
	if d := decide(t, heuristic.New(), in); !strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("with a blockable 6/6 on the third seat the race is safe; got %q", d.Reason)
	}

	rogue.Restrictions = []string{"cant_be_blocked"}
	v, mine = raceBoard(7, 6, 6, 6, rogue)
	in = input(0, v, attackAll(t, mine)...)
	if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("raced with an unblockable 6/6 pointed at a player on 6 (%s)", d.Reason)
	}
}

// TestRaceCountsTrampleOverAChumpBlocker: the bot's only ground is two
// Bears and the defender has a 7/7 trampler. A Bear in front of it
// stops 2 of the 7 — the crack-back check must count a trampler as
// held only by a blocker that absorbs all of it, or it calls a lethal
// turn safe.
func TestRaceCountsTrampleOverAChumpBlocker(t *testing.T) {
	flier := keywords("flying")
	var cards []protocol.CardView
	var mine []string
	for i := 0; i < 7; i++ {
		cards = append(cards, creature(cardID(10+i), 0, "Drake", 3, 3, flier))
		mine = append(mine, cardID(10+i))
	}
	for i := 0; i < 2; i++ {
		cards = append(cards, creature(cardID(20+i), 0, "Bear", 2, 2))
		mine = append(mine, cardID(20+i))
	}
	for i := 0; i < 6; i++ {
		cards = append(cards, creature(cardID(30+i), 1, "Drake", 3, 3, flier))
	}
	cards = append(cards, creature(cardID(40), 1, "Wurm", 7, 7, keywords("trample")))
	v := newView(
		[]protocol.PlayerView{newSeat(0, withLife(6)), newSeat(1, withLife(6))},
		withBattlefield(cards...),
		withTurn(20, 0, "declare_attackers"),
	)
	in := input(0, v, attackAll(t, mine)...)
	d := decide(t, heuristic.New(), in)
	if strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("raced with a 7/7 trampler facing two Bears at 6 life (%s)", d.Reason)
	}
}

// TestRaceKeepsTheReserveHome: once the race is on, a creature outside
// the swing is the reserve the crack-back check counted on, so after
// the planned attacker is declared the bot stops — it does not add the
// Wurms because an individual Wurm attack happens to look fine.
func TestRaceKeepsTheReserveHome(t *testing.T) {
	v, mine := raceBoard(7, 6, 6, 6)
	// The first Drake is already declared (attacking, tapped).
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == mine[0] {
			v.Battlefield.Cards[i].AttackingTarget = seatID(1).String()
			v.Battlefield.Cards[i].Tapped = true
		}
	}
	in := input(0, v, attackAll(t, mine[1:])...)
	d := decide(t, heuristic.New(), in)
	if d.Index >= 0 && d.Index < len(in.Moves) && in.Moves[d.Index].Kind == legal.KindAttack {
		t.Fatalf("declared %q (%s) after the race's swing was complete; the rest is the reserve",
			in.Moves[d.Index].Label, d.Reason)
	}
}

// indexOfCard finds the battlefield index of an attack move's attacker.
func indexOfCard(v protocol.GameView, m legal.Move) int {
	for i := range v.Battlefield.Cards {
		if v.Battlefield.Cards[i].InstanceID == m.Source.String() {
			return i
		}
	}
	return -1
}
