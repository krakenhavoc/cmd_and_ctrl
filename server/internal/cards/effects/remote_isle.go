package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Remote Isle — Land (EDHREC rank 4253):
//
//	"This land enters tapped.
//	 {T}: Add {U}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The blue member of the same Urza's Saga cycling-land cycle as
// Polluted Mire beside it. Registered together because half a cycle
// is worse than none.
//
// Cycling {2} arrived with #660, exactly as on Polluted Mire.
func init() {
	Register(Spec{
		OracleID:     "24aebda0-315f-4d2f-8bd9-00bbaf5bd76a",
		Name:         "Remote Isle",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{Cycling("{2}")},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
