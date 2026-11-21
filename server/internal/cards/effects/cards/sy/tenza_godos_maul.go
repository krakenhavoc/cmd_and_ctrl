package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tenza, Godo's Maul — Legendary Artifact — Equipment {3}:
//
//	"Equipped creature gets +1/+1. As long as it's legendary, it gets an
//	 additional +2/+2. As long as it's red, it has trample.
//	 Equip {1}"
//
// Champion's Helm's shape: the +1/+1 is unconditional (PumpAttached); the
// +2/+2 and the trample are conditions read off the HOST on every layer
// recompute, so they switch on and off as the host's supertypes and
// colours change, with nothing captured at equip time. The host is the
// permanent as the earlier layers left it this pass, which is where
// "legendary" (layer 4) and "red" (layer 5) are decided.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8180ed7-84e9-49f7-8256-dcc21f9d334b",
		Name:         "Tenza, Godo's Maul",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			{
				Label:    "As long as it's legendary, it gets an additional +2/+2.",
				Layer:    game.Layer7PT,
				SubLayer: game.SubLayer7C_Modify,
				AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
					return effects.AttachedToSource(target, g, source) && target.IsLegendary()
				},
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					c.Power += 2
					c.Toughness += 2
				},
			},
			effects.GrantToAttachedWhile(
				func(host *game.Card, _ *game.Game, _ *game.Card) bool {
					return host.HasColorWord("R")
				},
				"trample",
			),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
