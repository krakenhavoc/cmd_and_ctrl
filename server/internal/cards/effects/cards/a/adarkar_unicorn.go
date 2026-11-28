package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Adarkar Unicorn — Creature — Unicorn {1}{W}{W}, 2/2 (Ice Age, S82 OLD batch
// 851, #5636, seam #6189):
//
//	"{T}: Add {U} or {C}{U}. Spend this mana only to pay cumulative upkeep
//	 costs."
//
// "Add {U} or {C}{U}" is one tap yielding one of two amounts, so it is two mana
// abilities ({T}: Add {U} and {T}: Add {C}{U}): either needs the same {T}, so
// the player gets exactly what the printed choice gives. Both carry #6189's
// cumulative-upkeep spend restriction, so the mana pays an upkeep's payment
// prompt and nothing else (CR 106.6, CR 702.24a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "287cd5e8-51e2-42a4-a139-c112ff89706d",
		Name:         "Adarkar Unicorn",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{U}",
				Label:        "{T}: Add {U}. Spend this mana only to pay cumulative upkeep costs.",
				Restrictions: []string{game.ManaRestrictCumulativeUpkeep},
			},
			{
				Cost:         effects.ManaAbilityCost{Tap: true},
				Produced:     "{C}{U}",
				Label:        "{T}: Add {C}{U}. Spend this mana only to pay cumulative upkeep costs.",
				Restrictions: []string{game.ManaRestrictCumulativeUpkeep},
			},
		},
	})
}
