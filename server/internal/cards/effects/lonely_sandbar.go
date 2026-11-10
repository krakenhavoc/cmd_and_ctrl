package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lonely Sandbar — Land (Commander deck reprint):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 Cycling {U} ({U}, Discard this card: Draw a card.)"
//
// Canyon Slough's shape with one colour: enters tapped, one mana
// ability, and the row below.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "765863c8-1be0-4bb1-9e9c-db7701cffde3",
		Name:         "Lonely Sandbar",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{U}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
