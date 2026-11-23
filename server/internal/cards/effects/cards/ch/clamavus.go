package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Clamavus — Creature — Human Tyranid Artificer {4}{G}, 3/3:
//
//	"Proclamator Hailer — Each creature you control gets +1/+1 for each +1/+1
//	 counter on it."
//
// A layer 7c modification that reads each creature's own counters (the
// counters themselves are layer 7d and unaffected by the pump), so a creature
// with two counters is +2/+2 on top of them. It includes Clamavus itself.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d772b16b-31db-48aa-8bb6-da32f45eebe4",
		Name:         "Clamavus",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Each creature you control gets +1/+1 for each +1/+1 counter on it.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.CreaturesYouControl,
			Apply: func(c *game.Characteristic, target *game.Card, _ *game.Game, _ *game.Card) {
				n := target.Counters[game.CounterPlusOne]
				c.Power += n
				c.Toughness += n
			},
		}},
	})
}
