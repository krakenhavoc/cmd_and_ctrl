package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// trade_test.go is ADR 0104 owner decision 8: the bot trades Perplexing
// Chimera for a spell when the spell's mana value is 5 or more, or when
// it is a permanent spell — and keeps the creature otherwise.

const tradeChoiceID = "00000000-0000-4000-8000-00000000c104"

func tradeDecision(t *testing.T, stackSpell protocol.CardView) string {
	t.Helper()
	view := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withStack(stackSpell),
		withChoice(protocol.PendingChoiceView{
			ID: tradeChoiceID, Kind: "trigger_prompt", Chooser: seatID(0).String(),
			Reason:   "Perplexing Chimera — exchange control of this creature and that spell?",
			TradeFor: stackSpell.InstanceID,
		}))
	no, yes := false, true
	in := input(0, view,
		choiceMove(t, 0, tradeChoiceID, "keep the Chimera", map[string]any{"apply": no}),
		choiceMove(t, 0, tradeChoiceID, "exchange", map[string]any{"apply": yes}),
	)
	return chose(t, in, decide(t, heuristic.New(), in))
}

func TestChimeraTradesForABigSpell(t *testing.T) {
	if got := tradeDecision(t, spell(cardID(40), 1, "Expropriate", "{7}{U}{U}")); got != "exchange" {
		t.Errorf("a nine-mana spell: %q, want the exchange", got)
	}
}

func TestChimeraTradesForAPermanentSpell(t *testing.T) {
	bears := spell(cardID(41), 1, "Grizzly Bears", "{1}{G}")
	bears.TypeLine = "Creature — Bear"
	if got := tradeDecision(t, bears); got != "exchange" {
		t.Errorf("a creature spell: %q, want the exchange", got)
	}
}

func TestChimeraKeepsItselfOverACheapInstant(t *testing.T) {
	if got := tradeDecision(t, spell(cardID(42), 1, "Lightning Bolt", "{R}")); got != "keep the Chimera" {
		t.Errorf("a one-mana instant: %q, want to keep the Chimera", got)
	}
}
