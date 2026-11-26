package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// S82OldB169NotPutIntoGraveyardThisCombat passes a card that was not put into
// a graveyard (from any other zone) during the current combat: its last
// arrival there is not among the events since the combat phase's latest
// beginning-of-combat step (Storrev, Devkarin Lich: "that wasn't put there
// this combat"). Outside combat nothing was put there this combat, so every
// card passes.
func S82OldB169NotPutIntoGraveyardThisCombat(g *game.Game, _ uuid.UUID, c game.Card) bool {
	events := g.EventsThisTurn()
	start := -1
	for i, ev := range events {
		if ev.Kind == game.EventStepBegan && ev.Step == game.StepBeginCombat {
			start = i
		}
	}
	if start < 0 {
		return true
	}
	for _, ev := range events[start:] {
		if ev.CardID == c.InstanceID && ev.NewZone == game.ZoneGraveyard && ev.OldZone != game.ZoneGraveyard {
			return false
		}
	}
	return true
}
