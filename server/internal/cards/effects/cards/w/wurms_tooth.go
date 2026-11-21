package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wurm's Tooth — Artifact {2} (Darksteel, #5605):
//
//	"Whenever a player casts a green spell, you may gain 1 life."
//
// Angel's Feather's trigger with green: any player's cast counts, its
// controller's included (tmcB8APlayerCast). The colour is a word of the
// Tooth's text, read through it (CR 612.2). The trigger resolves even if
// the spell is countered in response. The life gain is a "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69476326-fda5-48c8-a506-e71eb7c81c2f",
		Name:         "Wurm's Tooth",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("G")),
				"Wurm's Tooth — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Wurm's Tooth — gain 1 life?"),
		},
	})
}
