package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// The Wanderer — Legendary Planeswalker {3}{W}, loyalty 5 (Commander
// Masters, #5948):
//
//	"Prevent all noncombat damage that would be dealt to you and other
//	 permanents you control.
//	 −2: Exile target creature with power 4 or greater."
//
// The prevention is two plain statics over noncombat damage, as Blessed
// Sanctuary's: one for the controller (ToYou) and one for every other
// permanent they control, of any type (ToOtherPermanentsYouControl, the
// Wanderer itself excluded: it takes the damage a creature's or spell's
// source deals to it). Combat damage is untouched. The −2 is Blade
// Banish's target and exile, the power read through the layers as it is
// announced and again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4c2196e5-6ed7-4cd2-b8e7-f01aff236569",
		Name:            "The Wanderer",
		StartingLoyalty: 5,
		Completeness:    effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToYou,
				Damage: effects.NoncombatDamage,
				Label:  "The Wanderer — prevent all noncombat damage that would be dealt to you",
			}),
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToOtherPermanentsYouControl,
				Damage: effects.NoncombatDamage,
				Label:  "The Wanderer — prevent all noncombat damage that would be dealt to other permanents you control",
			}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "−2: Exile target creature with power 4 or greater.",
			Cost:    effects.LoyaltyCost(-2),
			Targets: effects.TargetCreature("target creature with power 4 or greater", effects.PowerGE(4)),
			Effect:  effects.Do(effects.ExileChosenTarget{}),
		}},
	})
}
