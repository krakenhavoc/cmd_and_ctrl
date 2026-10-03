package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Carnelian Orb of Dragonkind — Artifact {2}{R}:
//
//	"{T}: Add {R}. If that mana is spent on a Dragon creature spell,
//	 it gains haste until end of turn."
//
// A mana ability with a spend rider that gives the SPELL haste until end
// of turn (ADR 0109 §11 decision 3, #1552), filtered on a Dragon
// creature spell. The grant rides the spell onto the permanent
// (CR 400.7a) and ends in the cleanup step. A changeling creature spell
// is a Dragon (CR 702.73a), as the filter vocabulary already reads it.
//
// One declared simplification, the rider's: with strict mana off the
// engine never spends the pool, so no rider fires (ADR 0068 §3).
func init() {
	Register(Spec{
		OracleID:     "651c967c-8f71-4eb5-b22f-545e55ea050e",
		Name:         "Carnelian Orb of Dragonkind",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so the Dragon never gains haste."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
			SpendRiders: []game.ManaSpendRider{
				SpentSpellGains([]string{"haste"}, true,
					ManaRestrictCast, ManaRestrictType("Creature"), ManaRestrictSubtype("Dragon")),
			},
		}},
	})
}
