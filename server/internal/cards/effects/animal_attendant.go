package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Animal Attendant — Creature — Human Citizen {1}{G}, 2/2 (Avatar: The
// Last Airbender Eternal, #5075):
//
//	"{T}: Add one mana of any color. If that mana is spent to cast a
//	 non-Human creature spell, that creature enters with an additional
//	 +1/+1 counter on it."
//
// Biophagus's mana and entry spend rider (#1547), with the spend filter
// narrowed to a non-Human creature spell: game.ManaRestrictNotSubtype
// reads the spell's subtypes as it is paid for, and a changeling spell is
// a Human (CR 702.73a), so it gets no counter. The counter is seeded on
// the creature's entry event, so a counter doubler sees it. The mana
// itself is unrestricted: spent on anything else it is ordinary mana.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "787e6cc6-d97a-4229-97d1-ec2377d138d3",
		Name:           "Animal Attendant",
		ReadsManaSpent: game.ReadsPoolManaOnly,
		Completeness:   CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
			SpendRiders: []game.ManaSpendRider{
				SpentEntersWithCounters("+1/+1", 1, ManaRestrictCast, ManaRestrictType("Creature"),
					game.ManaRestrictNotSubtype("Human")),
			},
		}},
	})
}
