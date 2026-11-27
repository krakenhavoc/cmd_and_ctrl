package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Terra Eternal — Enchantment {2}{W} (Worldwake, S82 OLD batch 490, #5582):
//
//	"All lands have indestructible."
//
// A layer 6 ability grant to every land on the battlefield, whoever controls
// it, as Darksteel Forge grants it to artifacts. Destroy effects, lethal
// damage and deathtouch no longer destroy them; sacrifice and state-based
// 0-toughness rules still do (CR 702.12b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a5ce8125-da02-48b0-83cd-871c934352b0",
		Name:         "Terra Eternal",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label: "All lands have indestructible.",
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, _ *game.Game, _ *game.Card) bool {
				return target.IsLand()
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				if !effects.KeywordSliceContains(c.Abilities, "indestructible") {
					c.Abilities = append(c.Abilities, "indestructible")
				}
			},
		}},
	})
}
