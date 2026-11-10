package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tranquil Thicket — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 Cycling {G} ({G}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, one mana
// ability, and the row below.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9f8fe514-77ed-41b4-a6f3-c6f095bb97be",
		Name:         "Tranquil Thicket",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{G}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
