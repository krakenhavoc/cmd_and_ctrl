package cardsw

import (
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Well-Laid Plans — Enchantment {2}{U} (Invasion):
//
//	"Prevent all damage that would be dealt to a creature by another
//	 creature if they share a color."
//
// A CR 615.1a prevention over damage to a creature battlefield permanent
// from a different creature source (read as last-known information, CR
// 608.2h) when the two share at least one colour (CR 105.2). Colourless
// creatures share a colour with nothing. Combat and noncombat damage alike.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50343947-d127-4479-9c1b-8f5040a97afb",
		Name:         "Well-Laid Plans",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{
			effects.PreventAllDamageDealtTo(effects.PreventionStatic{
				To:     effects.ToACreature,
				Damage: effects.AnyDamage,
				From:   wellLaidPlansFromAnotherCreatureSharingAColor(),
				Label:  "Well-Laid Plans — prevent all damage that would be dealt to a creature by another creature if they share a color",
			}),
		},
	})
}

func wellLaidPlansFromAnotherCreatureSharingAColor() effects.DamageSourceFilter {
	return func(ev *game.ReplacementEvent, g *game.Game, lki *game.Characteristic, _ *game.Card) bool {
		if lki == nil || !slices.Contains(lki.Types, "Creature") || ev.DamageSource == ev.DamageTarget {
			return false
		}
		target, ok := g.LookupCardForEffect(ev.DamageTarget)
		if !ok {
			return false
		}
		for _, c := range target.EffectiveColors() {
			if slices.Contains(lki.Colors, c) {
				return true
			}
		}
		return false
	}
}
