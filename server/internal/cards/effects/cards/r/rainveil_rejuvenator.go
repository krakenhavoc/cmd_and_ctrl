package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rainveil Rejuvenator — Creature — Elephant Druid {3}{G}, 2/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, you may mill three cards. (You may put the
//	 top three cards of your library into your graveyard.)
//	 {T}: Add an amount of {G} equal to this creature's power."
//
// The enters trigger is World Shaper's optional mill, asked as it
// resolves. The mana ability is Topiary Lecturer's (tdmGreenEqualToPower):
// the power as the ability is activated, layers brought current first, so a
// pump or a counter counts and a power of zero adds nothing. Summoning
// sickness applies to the tap (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6ccf72ee-a86a-4b13-9e76-e308d0d28523",
		Name:         "Rainveil Rejuvenator",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.WhenThisEnters("Rainveil Rejuvenator — mill three cards", effects.Do(effects.MillCards{N: 3})),
				"Rainveil Rejuvenator — mill three cards?"),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost:           effects.ManaAbilityCost{Tap: true},
			Label:          "Add an amount of {G} equal to this creature's power",
			ProducedAmount: effects.TdmGreenEqualToPower,
		}},
	})
}
