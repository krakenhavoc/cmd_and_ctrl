package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Supportive Parents — Creature — Human Citizen {2}{G}, 3/3 (Marvel's
// Spider-Man, #4460):
//
//	"Tap two untapped creatures you control: Add one mana of any color."
//
// Heritage Druid's mana ability with a count of two and Springleaf Drum's
// any-color production. The cost is not the {T} symbol, so summoning
// sickness does not stop a creature paying it, and Supportive Parents may
// tap itself as one of the two (it is an untapped creature you control).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b1266feb-516e-4b44-b750-7988ca6e6e55",
		Name:         "Supportive Parents",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost: ManaAbilityCost{TapOthers: &game.TapOthersCost{
				Count:  2,
				Filter: TargetPermanent("two untapped creatures you control", Creature()),
				Label:  "two untapped creatures you control",
			}},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
	})
}
