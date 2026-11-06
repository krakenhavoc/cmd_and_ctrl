package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dutiful Knowledge Seeker — Creature — Fox Spirit {2}{U}, 2/2 (Avatar:
// The Last Airbender Eternal, #5075):
//
//	"Whenever one or more cards are put into a library from anywhere,
//	 put a +1/+1 counter on this creature.
//	 {3}: Put target card from a graveyard on the bottom of its owner's
//	 library."
//
// The trigger is one per event batch, however many cards the batch put
// into libraries (avB29WheneverCardsArePutIntoALibrary). A scry or any
// other reordering of a library is not a card put into one, and a token
// is not a card. The activation is Chrome Companion's, with its own
// cost; it feeds the trigger.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b58a70e2-3609-441a-993a-dccb9a2cf242",
		Name:         "Dutiful Knowledge Seeker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			avB29WheneverCardsArePutIntoALibrary("Dutiful Knowledge Seeker — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		Activated: []ActivatedAbility{{
			Label:   "{3}: Put target card from a graveyard on the bottom of its owner's library.",
			Cost:    ManaCost("{3}"),
			Targets: TargetCardInGraveyard("target card from a graveyard"),
			Effect:  socB37PutTargetGraveyardCardOnBottom,
		}},
	})
}
