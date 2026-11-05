package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// trek_b24_helpers.go — shared bodies for the Star Trek batch 24 cards
// (#4973).

// trekB24TriggeringSpellIsSaga reports whether a cast trigger's spell is a
// Saga spell: its snapshot on the trigger as it was cast, else the card
// wherever it is now.
func trekB24TriggeringSpellIsSaga(g *game.Game, item *game.StackItem) bool {
	if item.Trigger == nil {
		return false
	}
	if obj := item.Trigger.Object; obj != nil && obj.ID == item.Trigger.Event.CardID {
		for _, s := range obj.Subtypes {
			if strings.EqualFold(s, "Saga") {
				return true
			}
		}
		return false
	}
	c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
	return ok && c.HasSubtype("Saga")
}
