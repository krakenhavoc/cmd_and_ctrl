package protocol

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trigger_order_why_view_test.go — #3015's wire half: the sentence under
// "Order your triggers" rides the trigger_order prompt as
// trigger_order_why, and the key is absent when there is no sentence.

func queueTriggerOrderPrompt(t *testing.T, g *game.Game, why string) *game.Player {
	t.Helper()
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:            game.PendingChoiceTriggerOrder,
			Chooser:         me.ID,
			Count:           2,
			Reason:          "Order your triggers",
			TriggerOrderIDs: []uuid.UUID{uuid.New(), uuid.New()},
			TriggerOrderWhy: why,
		})
	})
	return me
}

func TestTriggerOrderWhyRidesTheWire(t *testing.T) {
	g := newTwoSeatGame(t)
	me := queueTriggerOrderPrompt(t, g, "Ajani's Pridemate watches life totals.")

	v := FilterViewFor(ViewOfGame(g), me.ID.String())
	ch := choiceFor(t, v, string(game.PendingChoiceTriggerOrder))
	if ch.TriggerOrderWhy != "Ajani's Pridemate watches life totals." {
		t.Errorf("TriggerOrderWhy = %q", ch.TriggerOrderWhy)
	}
	raw, err := json.Marshal(ch)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), `"trigger_order_why":"Ajani's Pridemate watches life totals."`) {
		t.Errorf("the key is not on the wire: %s", raw)
	}

	// The table sees it too: it names public cards, like the heading.
	theirs := FilterViewFor(ViewOfGame(g), g.Seats[1].ID.String())
	if got := choiceFor(t, theirs, string(game.PendingChoiceTriggerOrder)).TriggerOrderWhy; got == "" {
		t.Error("an opponent's view lost the sentence")
	}
}

func TestTriggerOrderWhyIsOmittedWhenThereIsNone(t *testing.T) {
	g := newTwoSeatGame(t)
	me := queueTriggerOrderPrompt(t, g, "")

	ch := choiceFor(t, FilterViewFor(ViewOfGame(g), me.ID.String()), string(game.PendingChoiceTriggerOrder))
	raw, err := json.Marshal(ch)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "trigger_order_why") {
		t.Errorf("an empty sentence is on the wire: %s", raw)
	}
}

// Only a trigger_order prompt carries the field.
func TestTriggerOrderWhyIsOnlyOnTriggerOrderPrompts(t *testing.T) {
	g := newTwoSeatGame(t)
	me := g.Seats[0]
	g.WithWriteLock(func() {
		g.QueueChoiceForEffect(game.PendingChoice{
			Kind:            game.PendingChoiceDiscardFromHand,
			Chooser:         me.ID,
			FromPlayer:      me.ID,
			Count:           1,
			Reason:          "discard a card",
			TriggerOrderWhy: "should never be sent",
		})
	})
	ch := choiceFor(t, FilterViewFor(ViewOfGame(g), me.ID.String()), string(game.PendingChoiceDiscardFromHand))
	if ch.TriggerOrderWhy != "" {
		t.Errorf("a discard prompt carries %q", ch.TriggerOrderWhy)
	}
}
