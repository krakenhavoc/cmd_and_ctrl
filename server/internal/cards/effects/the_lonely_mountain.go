package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// The Lonely Mountain — Land — Mountain (The Hobbit, HOB):
//
//	"({T}: Add {R}.)
//	 This land enters tapped unless you control an Equipment.
//	 {4}{R}, {T}: Create a 2/2 red Dwarf creature token. This ability
//	 costs {1} less to activate for each Equipment you control. Activate
//	 only as a sorcery."
//
// Typed lands carry their mana as a row (the engine does not synthesise it
// for a nonbasic), and the tapped entry is the real CR 614 self-replacement,
// judged as the land enters. The token ability's own cost clause is Mobilized
// District's (ActivatedAbility.CostModifiers): {1} less for each Equipment
// you control as the cost is determined (CR 602.2b), floored at {R}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3678c06f-8a33-4a6d-bf20-5b92d5c05a95",
		Name:         "The Lonely Mountain",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedUnless(youControlAnEquipment)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label: "{4}{R}, {T}: Create a 2/2 red Dwarf creature token. This ability costs {1} less to activate for each Equipment you control. Activate only as a sorcery.",
			Cost:  Plus(ManaCost("{4}{R}"), TapCost()),
			CostModifiers: []game.CostModifier{
				CostsLessEach(PermanentsYouControl(HasSubtype("Equipment")),
					"This ability costs {1} less to activate for each Equipment you control."),
			},
			SorcerySpeed: true,
			Effect:       createTheToken("2/2 red Dwarf"),
		}},
	})
}

// youControlAnEquipment is the "unless you control an Equipment" test of
// The Lonely Mountain's tapped entry.
func youControlAnEquipment(g *game.Game, controller uuid.UUID) bool {
	for _, c := range g.BattlefieldCardsForEffect() {
		if c.Controller == controller && c.HasSubtype("Equipment") {
			return true
		}
	}
	return false
}
