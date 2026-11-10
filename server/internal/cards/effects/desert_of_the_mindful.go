package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert of the Mindful — Land — Desert:
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 Cycling {1}{U} ({1}{U}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, a mana ability
// declared by hand (a bare "Land — Desert" has no basic land type to derive
// one from, CR 305.6), and the shared Cycling constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "508f9e7e-2ff7-4593-b0a9-0612d7b5d646",
		Name:         "Desert of the Mindful",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{1}{U}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
