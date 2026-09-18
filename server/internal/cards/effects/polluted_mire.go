package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Polluted Mire — Land (EDHREC rank 4220):
//
//	"This land enters tapped.
//	 {T}: Add {B}.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The black member of Urza's Saga's cycling-land cycle. Nobody plays
// one for the mana — a tapped Swamp is a downgrade — they play it
// because a land that is a spell when you are flooded is worth a
// slot, and because it feeds a Life from the Loam or a Ramunap
// Excavator.
//
// DECLARED SIMPLIFICATION — NO CYCLING, the same one Ketria Triome
// and Raffine's Tower ship with, for the same two engine reasons:
//
//   - game.AbilityCost has no discard component. Spec.AdditionalCost's
//     DiscardCost is a spell-cast cost, not an ability cost.
//   - The CR 602 activation path only offers abilities on battlefield
//     permanents, and cycling is activated FROM HAND.
//
// That is the whole reason the card is played, so the caveat here is
// a much larger fraction of the card than it is on a Triome — a
// player sleeving this should know they are getting a tapped Swamp.
// Weaker than printed in the only direction we ship (#259); the
// tracker for the mechanic is #655.
func init() {
	Register(Spec{
		OracleID:     "9809d975-7ef8-4946-9041-607c4e954b13",
		Name:         "Polluted Mire",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Cycling {2} is not implemented — the card can only be played as a land, never cycled from hand, which is most of the reason the land is played."},
		Replacements: []game.ReplacementEffect{SelfEntersTapped()},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B}",
			Label:    "Add {B}",
		}},
	})
}
