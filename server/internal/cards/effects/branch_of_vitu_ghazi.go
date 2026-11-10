package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Branch of Vitu-Ghazi — Land (Murders at Karlov Manor, #4469):
//
//	"{T}: Add {C}.
//	 Disguise {3} (You may cast this card face down for {3} as a 2/2
//	 creature with ward {2}. Turn it face up any time for its disguise
//	 cost.)
//	 When this land is turned face up, add two mana of any one color.
//	 Until end of turn, you don't lose this mana as steps and phases
//	 end."
//
// The face-down permanent is the CR 708.2 2/2 creature with ward {2}; it
// has no abilities, so it taps for nothing until it is turned up. The
// turned-up trigger (CR 708.8) adds two mana of ONE colour (the pick is
// made once, OneColorOfAmount) carrying the keep-until-end-of-turn mark
// (KeepManaUntilEndOfTurn), the same one Savage Ventmaw adds.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7a30316b-dcd5-4a4b-b959-eecde7ca92e7",
		Name:             "Branch of Vitu-Ghazi",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Disguise("{3}")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp("Branch of Vitu-Ghazi — add two mana of any one color; you don't lose it as steps and phases end",
				Do(AddMana{
					Produced: OneColorOfAmount(2),
					Riders:   []game.ManaSpendRider{KeepManaUntilEndOfTurn()},
				})),
		},
	})
}
