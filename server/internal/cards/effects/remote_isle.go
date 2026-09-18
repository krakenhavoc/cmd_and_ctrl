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
// is worse than none: a player who has learned that Polluted Mire
// cannot be cycled here should not have to re-learn it per colour.
//
// DECLARED SIMPLIFICATION — NO CYCLING, exactly as on Polluted Mire;
// see that file for the two engine reasons and #655 for the tracker.
func init() {
	Register(Spec{
		OracleID:     "24aebda0-315f-4d2f-8bd9-00bbaf5bd76a",
		Name:         "Remote Isle",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Cycling {2} is not implemented — the card can only be played as a land, never cycled from hand, which is most of the reason the land is played."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
	})
}
