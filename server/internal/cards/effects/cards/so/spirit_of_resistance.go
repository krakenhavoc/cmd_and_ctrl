package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Spirit of Resistance — Enchantment {2}{W} (Invasion):
//
//	"As long as you control a permanent of each color, prevent all damage
//	 that would be dealt to you."
//
// A CR 615.1a prevention of all damage to the controller, combat or not,
// while the five colours are among the permanents they control (read as the
// damage would be dealt, through effective colours: a multicoloured
// permanent counts for each of its colours).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b075b645-2e5a-4d91-9440-02db4e6f5d08",
		Name:         "Spirit of Resistance",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.AnyDamage,
				While: func(g *game.Game, src *game.Card) bool {
					return effects.FiveColorsAmongYourPermanents(g, src.Controller, src.InstanceID)
				},
				Label: "Spirit of Resistance — as long as you control a permanent of each color, prevent all damage that would be dealt to you",
			}),
		},
	})
}
