package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mark of Asylum — Enchantment {1}{W} (Conflux, #5587):
//
//	"Prevent all noncombat damage that would be dealt to creatures you
//	 control."
//
// Blessed Sanctuary's second shield alone: a static prevention effect
// (CR 615.1a) over noncombat damage to the creatures its controller
// controls as the damage would be dealt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d4327e06-ee89-459c-a43a-dc064e198fef",
		Name:         "Mark of Asylum",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToCreaturesYouControl,
				Damage: effects.NoncombatDamage,
				Label:  "Mark of Asylum — prevent all noncombat damage that would be dealt to creatures you control",
			}),
		},
	})
}
