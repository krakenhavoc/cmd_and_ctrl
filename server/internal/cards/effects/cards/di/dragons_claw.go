package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dragon's Claw — Artifact {2} (Darksteel, #5605):
//
//	"Whenever a player casts a red spell, you may gain 1 life."
//
// Angel's Feather's trigger with red: any player's cast counts, its
// controller's included (tmcB8APlayerCast). The colour is a word of the
// Claw's text, read through it (CR 612.2). The trigger resolves even if
// the spell is countered in response. The life gain is a "you may"
// (CR 603.5).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd8d707d-1201-42c6-bce6-fbda12ce5ea8",
		Name:         "Dragon's Claw",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventCast, effects.TmcB8APlayerCast(effects.OfColor("R")),
				"Dragon's Claw — you may gain 1 life", effects.Do(effects.GainLife{Amount: 1})),
				"Dragon's Claw — gain 1 life?"),
		},
	})
}
