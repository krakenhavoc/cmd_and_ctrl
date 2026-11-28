package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cho-Manno, Revolutionary — Legendary Creature — Human Rebel {2}{W}{W}, 2/2 (Mercadian Masques, S82 OLD batch 688, #5619):
//
//	"Prevent all damage that would be dealt to Cho-Manno."
//
// A CR 615.1a prevention static whose recipient is the creature itself (Glittering Lion's shape), any damage from any source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91af5e35-b3b8-43ce-b1ea-997ed74e4ad2",
		Name:         "Cho-Manno, Revolutionary",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:    effects.ToThisCreature,
				Label: "Cho-Manno, Revolutionary — prevent all damage that would be dealt to Cho-Manno",
			}),
		},
	})
}
