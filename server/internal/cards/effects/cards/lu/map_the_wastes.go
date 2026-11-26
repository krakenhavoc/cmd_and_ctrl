package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Map the Wastes — {2}{G} Sorcery (Fate Reforged):
//
//	"Search your library for a basic land card, put it onto the battlefield
//	 tapped, then shuffle. Bolster 1."
//
// The bolster follows the shuffle, from the search's continuation, so it waits
// for a search prompt; the creature with the least toughness is chosen after
// the land has entered (CR 608.2c), a land that is also a creature included.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b845e69b-5d49-40ff-9cf7-d5cf2ccec7be",
		Name:         "Map the Wastes",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1, Answers: game.AnswerPump},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        item.Controller,
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				TappedOnEntry: true,
				Limit:         1,
				Shuffle:       true,
				Reason:        "Map the Wastes — a basic land card",
				Source:        item.SourceCardID,
				ThenTail:      ctx.Tail(s82OldB339MapBolster),
			}.Apply(ctx)
		},
	})
}

// s82OldB339MapBolster is "Bolster 1", after the search. An on-disk key: never
// renamed, never reused.
var s82OldB339MapBolster = effects.CardTail("then/s82-old-b339-map-the-wastes-bolster", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.Bolster{N: 1}.Apply(ctx)
})
