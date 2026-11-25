package protocol

import (
	"testing"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// pay_to_cancel_view_test.go — #6220: the "… unless they pay {N} before
// that step" payments waiting on a seat ride its own PlayerView
// (pay_to_cancel), and no other seat's frame carries them.

var viewPayToCancelBody = game.DelayedBody("test/view-pay-to-cancel", func(*game.Game, *game.StackItem, game.EffectParams) error { return nil })

func TestPayToCancelRowsAreTheSeatsOwn(t *testing.T) {
	g := busyTable(t, 1)
	payer := g.Seats[0]
	other := g.Seats[1]
	var id string
	g.WithWriteLock(func() {
		id = g.ScheduleDelayedTriggerForEffect(game.DelayedTrigger{
			Controller:  other.ID,
			Label:       "lose 1 life",
			At:          game.StepDraw,
			TurnOf:      payer.ID,
			Body:        viewPayToCancelBody,
			PayToCancel: &game.PayToCancel{Player: payer.ID, Cost: "{1}"},
		}).String()
	})
	seatOf := func(v GameView, idOf string) *PlayerView {
		for i := range v.Seats {
			if v.Seats[i].ID == idOf {
				return &v.Seats[i]
			}
		}
		return nil
	}
	own := seatOf(ViewOfGameFor(g, payer.ID.String()), payer.ID.String())
	if own == nil || len(own.PayToCancel) != 1 || own.PayToCancel[0].ID != id || own.PayToCancel[0].Cost != "{1}" || own.PayToCancel[0].At != string(game.StepDraw) {
		t.Fatalf("payer's own row = %+v", own)
	}
	if theirs := seatOf(ViewOfGameFor(g, other.ID.String()), payer.ID.String()); theirs == nil || len(theirs.PayToCancel) != 0 {
		t.Errorf("another seat sees the payer's rows: %+v", theirs)
	}
}
