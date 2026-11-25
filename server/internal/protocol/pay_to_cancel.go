package protocol

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// PayToCancelView is one "… unless they pay {N} before that step" payment
// waiting on a seat (#6220; game/pay_to_cancel.go): a row on the seat's own
// PlayerView, fired as `special_action {kind: "pay_to_cancel", card_id: id,
// auto_tap: true}` while the seat has priority.
type PayToCancelView struct {
	// ID is the waiting delayed trigger's ID; the action's card_id.
	ID string `json:"id"`
	// Cost is what paying costs, in brace notation ("{1}").
	Cost string `json:"cost"`
	// Label is what happens if it is not paid, as the trigger reads.
	Label string `json:"label"`
	// SourceName is the card the trigger is from ("Nafs Asp").
	SourceName string `json:"source_name,omitempty"`
	// At is the step it happens at the beginning of ("draw", "upkeep"):
	// the payment must be made before it.
	At string `json:"at"`
}

// viewOfPayToCancel projects the seat's waiting payments from the engine's
// own accessor. Nil for nearly every seat. Runs under the read lock
// ViewOfGame already holds.
func viewOfPayToCancel(g *game.Game, seat uuid.UUID) []PayToCancelView {
	var out []PayToCancelView
	for _, o := range g.PayToCancelOffersLocked(seat) {
		out = append(out, PayToCancelView{
			ID:         o.ID.String(),
			Cost:       o.Cost,
			Label:      o.Label,
			SourceName: o.SourceName,
			At:         string(o.At),
		})
	}
	return out
}
