package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tender Wildguide — Creature — Possum Druid {1}{G}, 2/2 (Bloomburrow,
// #4467):
//
//	"Offspring {2} (You may pay an additional {2} as you cast this spell.
//	 If you do, when this creature enters, create a 1/1 token copy of it.)
//	 {T}: Add one mana of any color.
//	 {T}: Put a +1/+1 counter on this creature."
//
// Offspring is the shared keyword (offspring.go). The mana ability is
// Birds of Paradise's, and the counter ability Phantom Nantuko's: both
// tap, so a summoning-sick Wildguide can use neither (CR 302.6), and
// the counter ability uses the stack while the mana ability does not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "74868bc0-6081-4f46-a2c6-9521cd5465aa",
		Name:          "Tender Wildguide",
		Completeness:  CompletenessFull,
		OptionalCosts: []game.AdditionalCost{Offspring("{2}")},
		Triggered:     []game.TriggeredAbility{OffspringToken("Tender Wildguide")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W|U|B|R|G}",
			Label:    "Add one mana of any color",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{T}: Put a +1/+1 counter on this creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    TapCost(),
			Effect:  putCounterOnSelf,
		}},
	})
}
