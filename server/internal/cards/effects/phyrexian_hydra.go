package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phyrexian Hydra — Creature — Phyrexian Hydra {3}{G}{G}, 7/7:
//
//	"Infect
//	 If damage would be dealt to this creature, prevent that damage. Put a -1/-1 counter on this creature for each 1 damage prevented this way."
//
// ADR 0108 §8 (#1906): a prevention static whose additional effect counts
// the damage PREVENTED this way: damage that can't be prevented is dealt
// as normal and puts on no -1/-1 counter (CR 615.12; the ruling). One
// application per recipient in a damage instance, with the instance's
// total.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:        "b16085d5-6d00-4d47-ab8b-d18d55c72141",
		Name:            "Phyrexian Hydra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Replacements: []game.ReplacementEffect{
			PreventDamageDealtTo(PreventionStatic{
				To:    ToThisCreature,
				Then:  phyrexianHydraCountersBody,
				Label: "Phyrexian Hydra — prevent damage to it and put a -1/-1 counter on it for each 1 prevented",
			}),
		},
	})
}

// The additional effect: a -1/-1 counter per 1 damage prevented.
var phyrexianHydraCountersBody = game.DelayedBody("phyrexian-hydra/minus-counters-per-prevented", phyrexianHydraCounters)

func phyrexianHydraCounters(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	if _, ok := followUpThis(g, item); !ok || p.Amount <= 0 {
		return nil
	}
	return g.AddCounterByForEffect(item.Controller, item.SourceCardID, game.CounterMinusOne, p.Amount)
}
