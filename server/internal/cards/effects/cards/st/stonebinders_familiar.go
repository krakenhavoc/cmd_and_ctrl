package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stonebinder's Familiar — Creature — Spirit Dog {W}, 1/1 (Strixhaven:
// School of Mages, #5502):
//
//	"Whenever one or more cards are put into exile during your turn, put a
//	 +1/+1 counter on this creature. This ability triggers only once each
//	 turn."
//
// One trigger however many cards move at once (OncePerBatch) and only the
// first of the turn (TriggersOnlyOnceEachTurn). Tokens are not cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87266dec-ac03-4f21-b235-67c14c3024db",
		Name:         "Stonebinder's Familiar",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggersOnlyOnceEachTurn(effects.OncePerBatch(effects.OnAny([]game.EventKind{game.EventZoneMove, game.EventLTB}, effects.S82OldB40CardsPutIntoExileOnYourTurn,
				"Stonebinder's Familiar — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})))),
		},
	})
}
