package heuristic_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// outside_cast_test.go — #3955: a bot that owns cards outside the game
// takes them for a free cast, and casts the ones it was granted.

func outsideCastMove(t *testing.T, seat int, id, label string) legal.Move {
	return legal.Move{
		Type: legal.TypeCastSpell, Player: seatID(seat), Kind: legal.KindCast,
		Label: label, Source: uuid.MustParse(id),
		Params: mustJSON(t, map[string]any{"instance_id": id, "from_zone": "outside", "auto_tap": true}),
	}
}

// A pick whose cards are cast free from outside the game is priced as a
// take, not a discard: the bot names the best cards rather than none.
func TestAFreeCastPickFromTheSideboardTakesTheBestCard(t *testing.T) {
	const choiceID = "choice-spawnsire"
	small := creature(cardID(1), 0, "Eldrazi Drone", 1, 1)
	big := creature(cardID(2), 0, "Ulamog's Crusher", 8, 8)
	big.ManaCost = "{8}"
	pick := handChoice(choiceID, "Spawnsire of Ulamog — choose any number of Eldrazi cards to cast", 0, 2, small, big)
	pick.ChooseDestination = "cast_free_outside"
	v := newView([]protocol.PlayerView{newSeat(0, withSideboard(small, big)), newSeat(1)},
		withBattlefield(land(cardID(30), 0)),
		withChoice(pick))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "choose nothing", ids()),
		choiceMove(t, 0, choiceID, "take the Drone", ids(cardID(1))),
		choiceMove(t, 0, choiceID, "take both", ids(cardID(1), cardID(2))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take both" {
		t.Fatalf("chose %q, want every Eldrazi taken for the free cast", got)
	}
}

// A creature granted a cast from outside the game is cast, not passed
// over: nothing in hand is spent on it.
func TestACreatureGrantedFromOutsideIsCast(t *testing.T) {
	big := creature(cardID(2), 0, "Ulamog's Crusher", 8, 8)
	big.ManaCost = "{8}"
	big.CastableHere = true
	v := newView([]protocol.PlayerView{newSeat(0, withSideboard(big)), newSeat(1)},
		withBattlefield(land(cardID(30), 0)))
	in := input(0, v,
		passMove(0),
		outsideCastMove(t, 0, cardID(2), "Cast Ulamog's Crusher from outside the game"),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "Cast Ulamog's Crusher from outside the game" {
		t.Fatalf("chose %q, want the granted creature cast from outside the game", got)
	}
}
