package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nissa's Expedition — Sorcery {4}{G} (Magic 2015, #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Search your library for up to two basic land cards, put them onto the
//	 battlefield tapped, then shuffle."
//
// Burnished Hart's search as a spell, with the convoke TapCost. The
// search has a stated quality, so it may find fewer than two.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "653010bc-45af-427e-a294-e0cc28d31821",
		Name:         "Nissa's Expedition",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Answers: game.AnswerValue, Lands: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     effects.IsBasicLand,
				Dest:          game.ZoneBattlefield,
				Limit:         2,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Nissa's Expedition — up to two basic lands, onto the battlefield tapped",
			}.Apply(ctx)
		},
	})
}
