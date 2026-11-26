package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Personal Sanctuary — Enchantment {2}{W} (Magic 2012, S82 OLD batch 452,
// #5573):
//
//	"During your turn, prevent all damage that would be dealt to you."
//
// Solitary Confinement's damage prevention (CR 615.1a), gated to its
// controller's own turn, read as the damage would be dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "20963cad-1c13-44a1-901a-bfc9930089e6",
		Name:         "Personal Sanctuary",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.AnyDamage,
				While:  func(g *game.Game, src *game.Card) bool { return effects.IsYourTurn(g, src.Controller) },
				Label:  "Personal Sanctuary — during your turn, prevent all damage that would be dealt to you",
			}),
		},
	})
}
