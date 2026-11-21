package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// boom_scholar_test.go — #1207 fixed the two root causes Boom
// Scholar shipped as declared caveats (charged-cost display, and the
// CR 601.2f pass reaching a CR 605 mana ability's own cost) but left
// clearing the card itself as a fast follow-up. This file is that
// follow-up's proof: the card's own CostModifier, unchanged since
// #1184, now reaches a MANA ability of another permanent because the
// engine plumbing under it does.
//
// boomScholarManaProbeOracle is a synthetic Signet-shaped exhaust
// mana ability ("{1}, {T}: Add {C}{C}") rather than Loot, the
// Pathfinder's printed "{G}, {T}": a reduction only ever spends
// GENERIC mana (CR 601.2f), and Loot's colored {G} can never show a
// discount landed. mana_ability_cost_modifier_test.go uses the same
// probe shape at the engine layer for the identical reason.
const BoomScholarManaProbeOracle = "test-boom-scholar-mana-probe"

func init() {
	effects.Register(effects.Spec{
		OracleID: BoomScholarManaProbeOracle,
		Name:     "Boom Scholar Mana Probe",
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true, Mana: "{1}"},
			Produced: "{C}{C}",
			Label:    "Exhaust — {1}, {T}: Add {C}{C}.",
			Exhaust:  true,
		}},
	})
}
