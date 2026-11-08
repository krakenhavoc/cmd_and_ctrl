package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Attuned Hunter — Creature — Human Ranger {2}{G}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Trample
//	 Whenever one or more cards leave your graveyard during your turn, put
//	 a +1/+1 counter on this creature."
//
// Spirit of Resilience's trigger (left_graveyard.go): every route a card
// leaves your graveyard by (an exile, a return, a cast from it, a land
// played from it) counts, and cards leaving together trigger it once
// (OncePerBatch, CR 603.2c). "During your turn" is part of the trigger
// condition, so cards that leave on an opponent's turn do nothing. A
// Hunter that has left the battlefield gets no counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f652c593-e8fa-42ff-a8b4-206556b273de",
		Name:            "Attuned Hunter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			OncePerBatch(OnAny(leftGraveyardEvents, whenCardsLeaveYourGraveyardDuringYourTurn,
				"Attuned Hunter — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1}))),
		},
	})
}
