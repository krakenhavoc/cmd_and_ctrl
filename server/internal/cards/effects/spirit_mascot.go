package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spirit Mascot — Creature — Spirit Ox {R}{W}, 2/2 (Secrets of Strixhaven,
// #4456):
//
//	"Whenever one or more cards leave your graveyard, put a +1/+1 counter on
//	 this creature."
//
// Ajani's Pridemate's counter on a "one or more" trigger: it watches the
// events that take a card out of a graveyard (leftGraveyardEvents, which
// count a cast or land play from a graveyard as well as a zone move) and
// reads the graveyard it left (whenCardsLeaveYourGraveyard), once per batch
// (OncePerBatch): several cards leaving at once are one trigger. A token
// ceasing to exist is not a card leaving.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34c572b2-dccf-43d9-a9c4-846653ad22da",
		Name:         "Spirit Mascot",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyard,
				"Spirit Mascot — put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
