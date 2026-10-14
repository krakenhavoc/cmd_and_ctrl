package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// sideboard_test.go — ADR 0191: a pick over the bot's own sideboard
// ("you may reveal a card you own from outside the game") brings the
// named cards in, so the bot takes the best one and never prices it as a
// discard.

func withSideboard(cards ...protocol.CardView) seatOpt {
	return func(p *protocol.PlayerView) {
		p.Sideboard = cards
		p.SideboardCount = len(cards)
	}
}

// outsideChoice is the prompt game.QueueOutsidePickForEffect queues.
func outsideChoice(id, reason string, lo, hi int, opts ...protocol.CardView) protocol.PendingChoiceView {
	c := handChoice(id, reason, lo, hi, opts...)
	c.ChooseDestination = "outside"
	return c
}

func TestAPickFromTheSideboardTakesTheBestCard(t *testing.T) {
	const choiceID = "choice-wish"
	small := creature(cardID(1), 0, "Bear", 2, 2)
	big := creature(cardID(2), 0, "Giant", 6, 6)
	big.ManaCost = "{4}{R}{R}"
	v := newView([]protocol.PlayerView{newSeat(0, withSideboard(small, big)), newSeat(1)},
		withBattlefield(land(cardID(30), 0)),
		withChoice(outsideChoice(choiceID, "Living Wish — reveal a creature or land card you own from outside the game, or none", 0, 1, small, big)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "choose nothing", ids()),
		choiceMove(t, 0, choiceID, "take the Bear", ids(cardID(1))),
		choiceMove(t, 0, choiceID, "take the Giant", ids(cardID(2))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "take the Giant" {
		t.Fatalf("chose %q, want the best card taken from outside the game", got)
	}
}

// "Exactly two with different names" (Extrapolate the Impossible): the
// bot reveals a pair rather than declining, whatever order the
// enumerator listed them in.
func TestAPickOfTwoFromTheSideboardRevealsAPair(t *testing.T) {
	const choiceID = "choice-extrapolate"
	a := creature(cardID(1), 0, "Bear", 2, 2)
	b := creature(cardID(2), 0, "Giant", 6, 6)
	b.ManaCost = "{4}{R}{R}"
	v := newView([]protocol.PlayerView{newSeat(0, withSideboard(a, b)), newSeat(1)},
		withBattlefield(land(cardID(30), 0)),
		withChoice(outsideChoice(choiceID, "Extrapolate the Impossible — reveal exactly two cards", 0, 2, a, b)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "choose nothing", ids()),
		choiceMove(t, 0, choiceID, "reveal both", ids(cardID(1), cardID(2))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "reveal both" {
		t.Fatalf("chose %q, want the pair revealed", got)
	}
}

// A pick whose candidates are not the bot's sideboard keeps its old
// price: the destination alone does not reprice a prompt.
func TestAnOutsideDestinationOverCardsNotInTheSideboardIsNotRepriced(t *testing.T) {
	const choiceID = "choice-odd"
	bear := creature(cardID(1), 0, "Bear", 2, 2)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(bear)), newSeat(1)},
		withBattlefield(land(cardID(30), 0)),
		withChoice(outsideChoice(choiceID, "Discard up to one card", 0, 1, bear)))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "choose nothing", ids()),
		choiceMove(t, 0, choiceID, "discard the Bear", ids(cardID(1))),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose nothing" {
		t.Fatalf("chose %q, want the hand pick left on its discard price", got)
	}
}
