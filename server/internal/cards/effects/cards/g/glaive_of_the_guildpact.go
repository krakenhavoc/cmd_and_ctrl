package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glaive of the Guildpact — Artifact — Equipment {2} (Guilds of Ravnica,
// #5525):
//
//	"Equipped creature gets +1/+0 for each Gate you control and has vigilance
//	 and menace.
//	 Equip {3}"
//
// Adaptive Omnitool's "for each permanent you control" pump counted over Gates
// (read live, through the layers, CR 613.4c, so a Gate that enters or leaves
// changes it at once), Sentinel's Eyes' keyword grant, and the shared Equip. The
// Gates are the equipment controller's, not the equipped creature's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "178a1b0b-30af-4afd-8976-4a0e2133007a",
		Name:         "Glaive of the Guildpact",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttachedPer(1, 0, func(g *game.Game, source *game.Card) int {
				return effects.CountPermanentsYouControl(g, source, func(c *game.Card) bool { return c.HasSubtypeWord("Gate") })
			}),
			effects.GrantToAttached("vigilance", "menace"),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}
