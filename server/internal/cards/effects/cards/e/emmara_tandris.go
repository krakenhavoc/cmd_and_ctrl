package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emmara Tandris — Legendary Creature — Elf Shaman {5}{G}{W}, 5/7
// (Commander 2019):
//
//	"Prevent all damage that would be dealt to creature tokens you
//	 control."
//
// Rune-Tail's Essence's prevention static (all damage, combat or not, from
// any source) with the recipient narrowed to the creature tokens its
// controller controls, read as the damage would be dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "38a71429-e673-4de4-80a8-f59e5cf98acc",
		Name:         "Emmara Tandris",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToCreatureTokensYouControl,
				Label: "Emmara Tandris — prevent all damage that would be dealt to creature tokens you control",
			}),
		},
	})
}
