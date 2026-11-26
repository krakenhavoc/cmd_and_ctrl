package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Retribution of the Ancients — Enchantment {B} (Khans of Tarkir):
//
//	"{B}, Remove X +1/+1 counters from among creatures you control: Target
//	 creature gets -X/-X until end of turn."
//
// Jetfire's variable-and-among counter cost (#3276): the activator announces
// the total and the split across their creatures as the cost is paid
// (CR 602.2b), and X is the total removed, read off the paid-cost record. The
// cost is a floor of one counter, as Jetfire's is: X of 0 would only give a
// creature -0/-0.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "550b8bee-1773-427b-b300-aaceff243e11",
		Name:         "Retribution of the Ancients",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}, Remove X +1/+1 counters from among creatures you control: Target creature gets -X/-X until end of turn.",
			Purpose: effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
			Cost: effects.Plus(effects.ManaCost("{B}"), game.AbilityCost{RemoveCounters: &game.CounterRemovalCost{
				Counter:  game.CounterPlusOne,
				N:        1,
				From:     effects.TargetPermanent("creatures you control", effects.Creature()),
				Among:    true,
				Variable: true,
			}}),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB350TargetGetsMinusXUntilEOT,
		}},
	})
}
