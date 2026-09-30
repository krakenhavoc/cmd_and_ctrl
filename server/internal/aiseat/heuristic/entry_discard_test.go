package heuristic_test

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/aiseat/heuristic"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// entry_discard_test.go — ADR 0098's two bot branches. The sign is the
// opposite of #1198's reveal: a card named to Mox Diamond's prompt is
// DISCARDED, and a permanent named to Heart of Yavimaya's is
// SACRIFICED, so the bot names the cheapest one — and for the Mox it
// still prefers the discard to losing the Mox, unless the land is one it
// cannot spare.

func TestEntryDiscardTakesTheMoxWithASpareLand(t *testing.T) {
	const choiceID = "choice-mox"
	a, b := land(cardID(1), 0), land(cardID(2), 0)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(a, b)), newSeat(1)},
		withBattlefield(land(cardID(10), 0), land(cardID(11), 0), land(cardID(12), 0)),
		withChoice(protocol.PendingChoiceView{
			ID:        choiceID,
			Kind:      "entry_discard_from_hand",
			Chooser:   seatID(0).String(),
			Reason:    "Mox Diamond — discard a land card so it enters?",
			Options:   []protocol.CardView{a, b},
			ChooseMin: 0,
			ChooseMax: 1,
		}))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "discard nothing", map[string]any{"card_ids": []string{}}),
		choiceMove(t, 0, choiceID, "discard land 1", map[string]any{"card_ids": []string{cardID(1)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard land 1" {
		t.Fatalf("chose %q, want the discard: a spare land for a Mox", got)
	}
}

func TestEntryDiscardKeepsTheOnlyLandWhenShort(t *testing.T) {
	const choiceID = "choice-mox-short"
	a := land(cardID(1), 0)
	v := newView([]protocol.PlayerView{newSeat(0, withHand(a)), newSeat(1)},
		withChoice(protocol.PendingChoiceView{
			ID:        choiceID,
			Kind:      "entry_discard_from_hand",
			Chooser:   seatID(0).String(),
			Reason:    "Mox Diamond — discard a land card so it enters?",
			Options:   []protocol.CardView{a},
			ChooseMin: 0,
			ChooseMax: 1,
		}))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "discard nothing", map[string]any{"card_ids": []string{}}),
		choiceMove(t, 0, choiceID, "discard land 1", map[string]any{"card_ids": []string{cardID(1)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "discard nothing" {
		t.Fatalf("chose %q, want to keep the only land with no lands in play", got)
	}
}

func TestEntrySacrificeAlwaysAnswers(t *testing.T) {
	const choiceID = "choice-heart"
	forest := land(cardID(1), 0)
	v := newView([]protocol.PlayerView{newSeat(0), newSeat(1)},
		withBattlefield(forest),
		withChoice(protocol.PendingChoiceView{
			ID:        choiceID,
			Kind:      "entry_sacrifice",
			Chooser:   seatID(0).String(),
			Reason:    "Heart of Yavimaya — sacrifice a Forest so it enters.",
			Options:   []protocol.CardView{forest},
			ChooseMin: 1,
			ChooseMax: 1,
		}))
	in := input(0, v,
		choiceMove(t, 0, choiceID, "sacrifice the Forest", map[string]any{"card_ids": []string{cardID(1)}}),
	)
	if got := chose(t, in, decide(t, heuristic.New(), in)); got != "sacrifice the Forest" {
		t.Fatalf("chose %q, want the only answer", got)
	}
}
