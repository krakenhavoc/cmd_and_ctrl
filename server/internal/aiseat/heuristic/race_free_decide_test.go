package heuristic_test

import (
	"strings"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// race_free_decide_test.go pins #2310 at the decision: the two-turn
// race has to beat the defender's free answer as well as the one that
// blocks to survive. A block is free when the blocker lives, or when
// it is the defender's commander, which is cast again (CR 903.9a).

// loopBoard is the board the full seed-1409 Wurm mirror looped on: both
// players on 6, my 3/3 commander and five Bears against their 3/3 and
// three Bears. Their 3/3 is their commander unless theirIsCommander is
// false. The moves are an attack by each of mine at them.
func loopBoard(t *testing.T, theirIsCommander bool) (protocol.GameView, []legal.Move) {
	t.Helper()
	cards := []protocol.CardView{creature(cardID(1), 0, "Chief", 3, 3, commander())}
	for i := 0; i < 5; i++ {
		cards = append(cards, creature(cardID(10+i), 0, "Bear", 2, 2))
	}
	theirs := creature(cardID(999), 1, "Chief", 3, 3)
	theirs.IsCommander = theirIsCommander
	cards = append(cards, theirs)
	for i := 0; i < 3; i++ {
		cards = append(cards, creature(cardID(20+i), 1, "Bear", 2, 2))
	}
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
	), attackAll(t, mine)
}

// declared marks one of the bot's creatures as already attacking seat 1,
// and drops its attack moves.
func declared(v protocol.GameView, moves []legal.Move, id string) (protocol.GameView, []legal.Move) {
	cards := append([]protocol.CardView(nil), v.Battlefield.Cards...)
	for i := range cards {
		if cards[i].InstanceID == id {
			cards[i].AttackingTarget = seatID(1).String()
			cards[i].Tapped = true
		}
	}
	v.Battlefield.Cards = cards
	var left []legal.Move
	for _, m := range moves {
		if m.Kind != legal.KindAttack || m.Source.String() != id {
			left = append(left, m)
		}
	}
	return v, left
}

// TestRaceWillNotBankDamageACommanderStopsForFree is the issue's loop.
// The race used to send my commander alone: "3 now + 4 next turn ≥
// their 6 life". The defender blocked it with its own commander, both
// were cast again, and the same race came round for a dozen turns.
// Against the free answer that swing is 0 now and 2 next. The smallest
// swing that races is my commander and two Bears — their commander can
// stop only one of the three for free — so with the commander declared
// the bot sends a Bear after it instead of stopping.
func TestRaceWillNotBankDamageACommanderStopsForFree(t *testing.T) {
	v, moves := loopBoard(t, true)
	in := input(0, v, moves...)
	d := decide(t, heuristic.New(), in)
	if !strings.Contains(d.Reason, raceLabel) || !strings.Contains(d.Reason, "4 now + 2 next turn") {
		t.Fatalf("decided %q (%s); want the race on my commander and two Bears, 4 now + 2 next turn", chose(t, in, d), d.Reason)
	}

	v, moves = declared(v, moves, cardID(1))
	in = input(0, v, moves...)
	d = decide(t, heuristic.New(), in)
	if got := chose(t, in, d); !strings.Contains(got, "Attack") {
		t.Fatalf("with my commander declared alone, decided %q (%s); their commander blocks it for free, so the swing needs the Bears", got, d.Reason)
	}
}

// TestRaceStillBanksDamageOnlyALostCreatureStops is the control: the
// same board with their 3/3 an ordinary creature. Blocking my commander
// with it now loses it for good, so the lone commander's 3 are damage
// the defender can stop only at a price, and the original race stands.
func TestRaceStillBanksDamageOnlyALostCreatureStops(t *testing.T) {
	v, moves := loopBoard(t, false)
	in := input(0, v, moves...)
	d := decide(t, heuristic.New(), in)
	if !strings.Contains(d.Reason, "3 now + 4 next turn") {
		t.Fatalf("decided %q (%s); want the lone commander's race, 3 now + 4 next turn", chose(t, in, d), d.Reason)
	}

	v, moves = declared(v, moves, cardID(1))
	in = input(0, v, moves...)
	d = decide(t, heuristic.New(), in)
	if got := chose(t, in, d); !strings.Contains(got, "Pass") {
		t.Fatalf("with my commander declared, decided %q (%s); the race on it alone is complete", got, d.Reason)
	}
}

// TestRaceCountsWhatTheFreeBlockKills: my 3/3 is the only creature that
// can attack, and a 5/5 that came down this turn waits at home. Their
// 3/3 commander is all they have, and they are on 3. The defender that
// blocks to survive takes the 3; the free answer trades its commander
// for my 3/3, casts it again, and next turn chumps the 5/5 with it — so
// the race is 0 now and 0 next, not the 0 now and 3 next it reads if
// the 3/3 is still counted alive.
func TestRaceCountsWhatTheFreeBlockKills(t *testing.T) {
	v, moves := smallBoard(t, 10, 3,
		creature(cardID(1), 0, "Ogre", 3, 3),
		creature(cardID(2), 0, "Giant", 5, 5, sick()),
		creature(cardID(999), 1, "Chief", 3, 3, commander()),
	)
	// The Giant is summoning sick: no attack for it.
	moves = moves[:2]
	in := input(0, v, moves...)
	if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("decided %q (%s); their commander trades with the 3/3 for free and chumps the Giant next turn", chose(t, in, d), d.Reason)
	}
}

// TestRaceFreeAnswersCommanderCannotSwingBack: my Ogre into a defender
// on 4 whose only creature is its 3/3 commander, me on 3 with two
// tapped Drakes. The free answer blocks the Ogre with the commander,
// which dies and is cast again — on the defender's own turn, too new to
// attack me (CR 302.6), so the crack-back is nothing and the Drakes
// finish it next turn. A commander with haste can swing back, and 3
// kill me.
func TestRaceFreeAnswersCommanderCannotSwingBack(t *testing.T) {
	for _, tc := range []struct {
		name  string
		opts  []cardOpt
		races bool
	}{
		{"cast again, it is too new to attack", []cardOpt{commander()}, true},
		{"with haste it attacks the turn it is cast", []cardOpt{commander(), keywords("haste")}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v, moves := smallBoard(t, 3, 4,
				creature(cardID(1), 0, "Ogre", 4, 4),
				creature(cardID(2), 0, "Drake", 3, 3, keywords("flying"), tapped()),
				creature(cardID(3), 0, "Drake", 3, 3, keywords("flying"), tapped()),
				creature(cardID(999), 1, "Chief", 3, 3, tc.opts...),
			)
			// The Drakes are tapped: no attacks for them.
			moves = moves[:2]
			in := input(0, v, moves...)
			d := decide(t, heuristic.New(), in)
			if got := strings.Contains(d.Reason, raceLabel); got != tc.races {
				t.Fatalf("decided %q (%s); raced = %v, want %v", chose(t, in, d), d.Reason, got, tc.races)
			}
		})
	}
}

// TestRaceCountsTheFreeAnswersCrackBack: my Wurm into a defender on 6
// with a 1/3 commander and an Ogre, me on 4 with two tapped Drakes and
// nothing at home. Blocking to survive, the defender chumps with the
// Ogre, which soaks the most of the Wurm, and only its 1-power
// commander can swing back. But it can block with the commander for
// free instead and keep the Ogre: the Wurm and the Drakes still have 6
// for it next turn, and the Ogre's 4 kill me first. The race has to
// survive that crack-back too.
func TestRaceCountsTheFreeAnswersCrackBack(t *testing.T) {
	v, moves := smallBoard(t, 4, 6,
		creature(cardID(1), 0, "Wurm", 7, 7, keywords("trample")),
		creature(cardID(2), 0, "Drake", 3, 3, keywords("flying"), tapped()),
		creature(cardID(3), 0, "Drake", 3, 3, keywords("flying"), tapped()),
		creature(cardID(999), 1, "Chief", 1, 3, commander()),
		creature(cardID(20), 1, "Ogre", 4, 4),
	)
	// The Drakes are tapped: no attacks for them.
	moves = moves[:2]
	in := input(0, v, moves...)
	if d := decide(t, heuristic.New(), in); strings.Contains(d.Reason, raceLabel) {
		t.Fatalf("decided %q (%s); the free answer keeps the Ogre, and it swings back for 4 at my 4", chose(t, in, d), d.Reason)
	}
}
