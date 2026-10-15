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
// No simplification is left (#3844, ADR 0068 amendment 2026-10-15). A
// waived payment (strict mana off with too little in the pool, or Cast
// anyway) still spends the share of the cost the pool can pay, so this
// mana pays whenever it can and its rider fires. Mana that never
// entered the pool was not this card's mana.
func init() {
	Register(Spec{
		OracleID:       "651c967c-8f71-4eb5-b22f-545e55ea050e",
		Name:           "Carnelian Orb of Dragonkind",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
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
