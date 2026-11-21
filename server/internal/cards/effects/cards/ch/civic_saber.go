package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Civic Saber — Artifact — Equipment {1} (Ravnica Remastered, #5948):
//
//	"Equipped creature gets +1/+0 for each of its colors.
//	 Equip {1}"
//
// A layer 7c static that reads the colours of the creature as built:
// colours are settled in layer 5, before layer 7 (CR 613.1), so a
// two-colour creature gets +2/+0 and a colourless one gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79441673-6e43-41bc-936c-958dbb9d92da",
		Name:         "Civic Saber",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Equipped creature gets +1/+0 for each of its colors.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				c.Power += len(c.Colors)
			},
		}},
		Activated: []effects.ActivatedAbility{
			effects.EquipAbility("{1}"),
		},
	})
}
