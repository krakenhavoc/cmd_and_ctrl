package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Barren Moor — Land:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Cycling {B} ({B}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with a single-colour mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "326ba371-124c-4949-a048-3a0c8962e567",
		Name:         "Barren Moor",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		Activated:    []ActivatedAbility{Cycling("{B}")},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
