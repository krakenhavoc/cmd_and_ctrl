package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert of the True — Land — Desert:
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 Cycling {1}{W} ({1}{W}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, a mana ability
// declared by hand (a bare "Land — Desert" has no basic land type to derive
// one from, CR 305.6), and the shared Cycling constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2672e0ca-8d5c-449f-8483-35d0e697fbb2",
		Name:         "Desert of the True",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{1}{W}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
