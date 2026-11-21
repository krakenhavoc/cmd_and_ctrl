package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Academy Raider — Creature — Human Warrior {2}{R}, 1/1 (Magic 2014, #5563):
//
//	"Intimidate (This creature can't be blocked except by artifact creatures and/or creatures that share a color with it.)
//	 Whenever this creature deals combat damage to a player, you may discard a card. If you do, draw a card."
//
// Fissure Wizard's rummage on combat damage to a player: the discard is
// optional, and only a card discarded draws one.
//
// No simplification.
func init() {
	const label = "Academy Raider — you may discard a card. If you do, draw a card"
	effects.Register(effects.Spec{
		OracleID:        "75131d75-0703-44d0-b503-35190be8e66f",
		Name:            "Academy Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"intimidate"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			}),
		},
	})
}
