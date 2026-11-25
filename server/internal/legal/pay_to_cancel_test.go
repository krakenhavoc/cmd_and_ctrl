package legal_test

import (
	"encoding/json"
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// pay_to_cancel_test.go — the enumerator half of #6220: a "… unless they
// pay {N} before that step" payment waiting on the seat is a special-action
// move while the seat has priority and can pay, and the dispatcher accepts
// it (#544).

var legalPayToCancelBody = game.DelayedBody("test/legal-pay-to-cancel", func(*game.Game, *game.StackItem, game.EffectParams) error { return nil })

func TestPayToCancelIsAMoveWhenAffordable(t *testing.T) {
	g := newTable(t)
	me := g.Seats[g.Turn.ActiveSeat]
	other := g.Seats[(g.Turn.ActiveSeat+1)%len(g.Seats)]
	advanceTo(t, g, game.StepPrecombatMain)
	var id string
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:  other.ID,
			Label:       "lose 1 life",
			At:          game.StepDraw,
			TurnOf:      me.ID,
			Body:        legalPayToCancelBody,
			PayToCancel: &game.PayToCancel{Player: me.ID, Cost: "{1}"},
		}).String()
	})
	find := func() []legal.Move {
		var out []legal.Move
		for _, m := range legal.EnumerateFor(g, me.ID) {
			var p struct {
				CardID string `json:"card_id"`
				Kind   string `json:"kind"`
			}
			if m.Type == legal.TypeSpecialAction && json.Unmarshal(m.Params, &p) == nil && p.Kind == "pay_to_cancel" && p.CardID == id {
				out = append(out, m)
			}
		}
		return out
	}
	clearHand(me)
	if got := find(); len(got) != 0 {
		t.Fatalf("offered with no mana: %v", labels(got))
	}
	battlefieldCard(g, me, game.Card{Name: "Mountain", TypeLine: "Basic Land — Mountain"})
	got := find()
	if len(got) != 1 {
		t.Fatalf("want one pay_to_cancel move, got %v", labels(got))
	}
	if moves := legal.EnumerateFor(g, other.ID); len(moves) > 0 {
		for _, m := range moves {
			if m.Type == legal.TypeSpecialAction && m.Label == got[0].Label {
				t.Errorf("another seat is offered the payment")
			}
		}
	}
	dispatchAll(t, g, me.ID, got)
}
