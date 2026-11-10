package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Desert of the Glorified — Land — Desert:
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Cycling {1}{B} ({1}{B}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, a mana ability
// declared by hand (a bare "Land — Desert" has no basic land type to derive
// one from, CR 305.6), and the shared Cycling constructor.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "190e664b-9875-4335-af46-353886ed18ff",
		Name:         "Desert of the Glorified",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{1}{B}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
