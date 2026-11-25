package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Arisen Gorgon — Creature — Zombie Gorgon {1}{B}{B}, 3/3 (Core Set 2019, S82 OLD batch 199, #5527):
//
//	"This creature has deathtouch as long as you control a Liliana
//	 planeswalker."
//
// Court Cleric's condition (a planeswalker with the Liliana subtype under your
// control) as a layer 6 keyword grant to itself.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "69858691-1190-4454-af0a-87c2ac4ab6ec",
		Name:         "Arisen Gorgon",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Has deathtouch as long as you control a Liliana planeswalker.",
			Layer:     game.Layer6Ability,
			AppliesTo: effects.SelfWhileYouControlA(effects.And(effects.Planeswalker(), effects.Subtype("Liliana"))),
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				effects.AppendKeywordsTo(c, []string{"deathtouch"})
			},
		}},
	})
}
