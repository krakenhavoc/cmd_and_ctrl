package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bog Gnarr — Creature — Beast {4}{G}, 2/2 (Apocalypse, S82 OLD, #5614):
//
//	"Whenever a player casts a black spell, this creature gets +2/+2 until end
//	 of turn."
//
// Glade Gnarr's trigger over black spells.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "537ebea5-f805-4f78-9151-58eb43c4a62d",
		Name:         "Bog Gnarr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("B")), "Bog Gnarr — +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Bog Gnarr — +2/+2 until end of turn")),
		},
	})
}
