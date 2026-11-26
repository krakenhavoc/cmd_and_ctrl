package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stoneforge Masterwork — Artifact — Equipment {1} (Oath of the Gatewatch,
// #5545):
//
//	"Equipped creature gets +1/+1 for each other creature you control that
//	 shares a creature type with it.
//	 Equip {2}"
//
// A layer 7c modify (CR 613.4c) that counts, on every recompute, the OTHER
// creatures the Masterwork's controller controls that share a creature type
// with the equipped creature (game.SharesCreatureType, so a changeling shares
// with everything, CR 702.73a). "You" is the Masterwork's controller, not the
// host's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83db5c28-e172-4a4d-9324-ed5ade5255bc",
		Name:         "Stoneforge Masterwork",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Equipped creature gets +1/+1 for each other creature you control that shares a creature type with it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.AttachedToSource,
			Apply: func(c *game.Characteristic, host *game.Card, g *game.Game, source *game.Card) {
				n := effects.OldB300SharedTypeCreaturesYouControl(g, host, source)
				c.Power += n
				c.Toughness += n
			},
		}},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
