package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// choose_source_test.go is ADR 0107 §6 decision 4: a bot asked for "a
// source of your choice" shields against the source most likely to deal
// the damage — an opponent's, a spell on the stack first, then the most
// power — and never its own.

const chooseSourceChoiceID = "00000000-0000-4000-8000-00000000e107"

func chooseSourceMoves(t *testing.T, ids ...string) []legal.Move {
	t.Helper()
	out := make([]legal.Move, 0, len(ids))
	for _, id := range ids {
		out = append(out, choiceMove(t, 0, chooseSourceChoiceID, "choose "+id, map[string]any{"card_ids": []string{id}}))
	}
	return out
}

func withChooseSource(cards ...protocol.CardView) viewOpt {
	return withChoice(protocol.PendingChoiceView{
		ID: chooseSourceChoiceID, Kind: "choose_source", Chooser: seatID(0).String(),
		FromPlayer: seatID(0).String(), Count: 1, ChooseMin: 1, ChooseMax: 1,
		Reason: "Circle of Protection: Red — choose a red source", Options: cards,
	})
}

func TestChooseSourcePrefersTheStrongestOpponentsCreature(t *testing.T) {
	mine := creature(cardID(1), 0, "My Dragon", 9, 9)
	small := creature(cardID(2), 1, "Goblin", 1, 1)
	big := creature(cardID(3), 2, "Giant", 6, 6)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1), newSeat(2)},
		withBattlefield(mine, small, big), withChooseSource(mine, small, big))
	in := input(0, v, chooseSourceMoves(t, cardID(1), cardID(2), cardID(3))...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose "+cardID(3) {
		t.Errorf("chose %q, want the opponent's 6-power Giant", got)
	}
}

func TestChooseSourcePrefersASpellOnTheStack(t *testing.T) {
	big := creature(cardID(3), 1, "Giant", 6, 6)
	bolt := spell(cardID(4), 1, "Lightning Bolt", "{R}")
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(big), withStack(bolt), withChooseSource(big, bolt))
	in := input(0, v, chooseSourceMoves(t, cardID(3), cardID(4))...)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "choose "+cardID(4) {
		t.Errorf("chose %q, want the Lightning Bolt about to resolve", got)
	}
}
