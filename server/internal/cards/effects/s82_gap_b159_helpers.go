package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b159_helpers.go — shared bodies of the GAP batch 159 cards (#5948).

// GapB159CyclesOfCardNamed counts the times `player` has cycled a card named
// `name` so far this game, as the engine logged them (one EventCycle per
// cycling activation). The card's name is read off the logged card.
func GapB159CyclesOfCardNamed(g *game.Game, player uuid.UUID, name string) int {
	n := 0
	for _, ev := range g.Events {
		if ev.Kind != game.EventCycle || ev.Actor != player {
			continue
		}
		if c, ok := g.LookupCardForEffect(ev.CardID); ok && c.Name == name {
			n++
		}
	}
	return n
}
