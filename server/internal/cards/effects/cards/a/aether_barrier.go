package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aether Barrier — Enchantment {2}{U} (Nemesis, S82 OLD, #5618):
//
//	"Whenever a player casts a creature spell, that player sacrifices a
//	 permanent of their choice unless they pay {1}."
//
// Isolation Cell's ask: the caster, whoever they are, pays {1} or sacrifices a
// permanent of their choice (a player with none sacrifices nothing).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0fe602b7-9f88-4d3d-af24-7790df867ed5",
		Name:         "Aether Barrier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCast, effects.OldB607CastACreatureSpell,
				"Aether Barrier — that player sacrifices a permanent of their choice unless they pay {1}",
				effects.OldB678SacrificeUnlessPay("Aether Barrier")),
		},
	})
}
