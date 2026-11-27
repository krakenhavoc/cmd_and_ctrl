package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glade Gnarr — Creature — Beast {5}{G}, 4/4 (Apocalypse, S82 OLD, #5614):
//
//	"Whenever a player casts a blue spell, this creature gets +2/+2 until end
//	 of turn."
//
// Angel's Feather's "a player casts a <colour> spell" (any player's cast, its
// controller's included), with the shared this-gets-until-end-of-turn body. The
// colour is a word of the Gnarr's text (CR 612.2). The trigger resolves even if
// the spell is countered in response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a42fd1d-a2c3-40c4-8c34-7a3559aa9caf",
		Name:         "Glade Gnarr",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("U")), "Glade Gnarr — +2/+2 until end of turn",
				effects.ThisGetsUntilEndOfTurn(2, 2, "Glade Gnarr — +2/+2 until end of turn")),
		},
	})
}
