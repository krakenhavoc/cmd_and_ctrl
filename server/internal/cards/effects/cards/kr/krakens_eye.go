package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kraken's Eye — Artifact {2} (Darksteel, #5605):
//
//	"Whenever a player casts a blue spell, you may gain 1 life."
//
// Angel's Feather's trigger with blue: any player's cast counts, its
// controller's included (tmcB8APlayerCast). The colour is a word of the
// Eye's text, read through it (CR 612.2). The trigger resolves even if
// the spell is countered in response. The life gain is a "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14a21d8a-2097-4150-8af5-8c155049477a",
		Name:         "Kraken's Eye",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("U")),
				"Kraken's Eye — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Kraken's Eye — gain 1 life?"),
		},
	})
}
