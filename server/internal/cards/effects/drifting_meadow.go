package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drifting Meadow — Land (Urza's Saga):
//
//	"This land enters tapped.
//	 {T}: Add {W}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The white member of Remote Isle's Urza's Saga cycling-land cycle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c6eb2814-0021-4308-ad44-6c8cc59b0d1c",
		Name:         "Drifting Meadow",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{W}",
			Label:    "Add {W}",
		}},
	})
}
