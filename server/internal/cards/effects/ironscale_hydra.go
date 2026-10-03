package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ironscale Hydra — Creature — Hydra {3}{G}{G}, 5/5:
//
//	"If a creature would deal combat damage to this creature, prevent that damage and put a +1/+1 counter on this creature."
//
// ADR 0108 §8 (#1906): "if a creature would deal" names the SOURCE, so the
// additional effect runs once per source in a damage instance: one +1/+1
// counter per creature whose combat damage is prevented, however much it
// was. Damage that can't be prevented is dealt and still adds the counter
// before lethal damage is checked (CR 615.12; the rulings). The source is
// judged a creature as the damage would be dealt.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cef72b9b-91b1-47ff-ab6e-91d3c548e98b",
		Name:         "Ironscale Hydra",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{
			PreventDamageASourceWouldDeal(PreventionStatic{
				To:     ToThisCreature,
				Damage: CombatDamage,
				From:   FromACreature(),
				Then:   aCounterOnThisBody,
				Label:  "Ironscale Hydra — prevent a creature's combat damage to it and put a +1/+1 counter on it",
			}),
		},
	})
}
