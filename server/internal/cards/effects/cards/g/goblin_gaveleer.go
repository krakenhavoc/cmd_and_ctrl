package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Goblin Gaveleer — Creature — Goblin Warrior {R}, 1/1 (Scars of Mirrodin, S82
// OLD batch 472, #5577):
//
//	"Trample
//	 This creature gets +2/+0 for each Equipment attached to it."
//
// Bruenor Battlehammer's per-Equipment pump on itself alone: a layer 7c +2/+0
// per Equipment attached to it, whoever controls the Equipment, read on every
// recompute. Trample is printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "739621e6-e138-4ef9-87e8-c730b5a7f495",
		Name:            "Goblin Gaveleer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Static: []game.StaticAbility{{
			Label:     "This creature gets +2/+0 for each Equipment attached to it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, target *game.Card, g *game.Game, _ *game.Card) {
				c.Power += 2 * effects.S82AfrB08EquipmentAttachedTo(g, target.InstanceID)
			},
		}},
	})
}
