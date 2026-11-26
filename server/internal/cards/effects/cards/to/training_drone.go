package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Training Drone — Artifact Creature — Drone {3}, 4/4 (Mirrodin Besieged, S82
// OLD b468, #5576):
//
//	"This creature can't attack or block unless it's equipped."
//
// A layer 6 self static that sets can't-attack and can't-block while no
// Equipment is attached to the Drone (CR 301.5; the check is Auriok
// Steelshaper's, read live on every recompute), so it may attack and block the
// moment one is attached and loses both as soon as the last leaves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4b40ad5f-b270-44e9-b62f-e94912bc3346",
		Name:         "Training Drone",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "This creature can't attack or block unless it's equipped.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return effects.SelfOnly(target, g, source) && !effects.B41SourceIsEquipped(g, source)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Restrictions |= game.CantAttack | game.CantBlock
			},
		}},
	})
}
