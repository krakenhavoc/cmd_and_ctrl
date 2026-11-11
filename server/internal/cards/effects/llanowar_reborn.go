package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Llanowar Reborn — Land (#5368):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 Graft 1 (This land enters with a +1/+1 counter on it. Whenever another
//	 creature enters, you may move a +1/+1 counter from this land onto it.)"
//
// Graft is cmdr2B23Graft (CR 702.58a), written once for this card: the
// counter is an entry replacement folded with "enters tapped" into one (so
// there is no ordering prompt) and the move is an optional trigger with an
// intervening "if" for any creature entering. The land is not a creature
// itself, so the counter sits on it harmlessly until moved.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "92acb789-0e42-465c-ac16-40fefec48805",
		Name:         "Llanowar Reborn",
		Completeness: CompletenessFull,
		Replacements: []game.ReplacementEffect{SelfEntersTappedWithCounters(game.CounterPlusOne, 1)},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Triggered: []game.TriggeredAbility{cmdr2B23Graft("Llanowar Reborn")},
	})
}
