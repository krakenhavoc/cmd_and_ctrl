package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secluded Steppe — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 Cycling {W} ({W}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, one mana
// ability, and the row below.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8dec6fcf-1254-4b1b-ba23-7a3e492a7241",
		Name:         "Secluded Steppe",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{W}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
