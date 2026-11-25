package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hedron Matrix — Artifact — Equipment {4} (Conspiracy: Take the Crown, S82
// older sets batch 273, #5540):
//
//	"Equipped creature gets +X/+X, where X is its mana value.
//	 Equip {4}"
//
// A layer 7c modification (CR 613.4c) of the equipped creature by its own
// mana value, read as the layers are applied, so it follows a creature that
// is copied or changes. A token that copies nothing has mana value 0, and X
// in a mana cost counts as 0 on the battlefield (CR 202.3e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dc748941-fd91-4a0e-b0c4-04c9092ff158",
		Name:         "Hedron Matrix",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Equipped creature gets +X/+X, where X is its mana value.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
				if target == nil {
					return
				}
				mv := target.ManaValue()
				c.Power += mv
				c.Toughness += mv
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}
