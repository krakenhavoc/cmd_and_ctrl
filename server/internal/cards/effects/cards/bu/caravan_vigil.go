package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Caravan Vigil — Sorcery {G} (Innistrad, S82 OLD batch 450, #5572):
//
//	"Search your library for a basic land card, reveal it, put it into
//	 your hand, then shuffle.
//	 Morbid — You may put that card onto the battlefield instead of
//	 putting it into your hand if a creature died this turn."
//
// Morbid is read as the spell resolves. When it holds, the "may" is asked
// before the search: the card is any basic land the controller picks, so
// nothing the choice depends on is learned from the search (the answer
// picks the destination, onto the battlefield untapped or into the hand).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "925a3ded-b7bc-4384-8684-79c20c59ab12",
		Name:         "Caravan Vigil",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if !effects.CondACreatureDiedThisTurn.Holds(ctx) {
				return caravanVigilSearch(ctx, game.ZoneHand)
			}
			return effects.MayChoice{
				Question: "Caravan Vigil — morbid: put that basic land onto the battlefield instead of into your hand?",
				Then:     caravanVigilMorbid,
			}.Apply(ctx)
		},
	})
}

var caravanVigilMorbid = effects.MayYesNo("confirm/s82-old-b450-caravan-vigil-morbid",
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return caravanVigilSearch(ctx, game.ZoneBattlefield)
	},
	func(ctx *effects.Context, _ game.ConfirmAnswer) error {
		return caravanVigilSearch(ctx, game.ZoneHand)
	})

func caravanVigilSearch(ctx *effects.Context, dest game.ZoneKind) error {
	return effects.SearchLibrary{
		Player:    ctx.Controller(),
		Predicate: effects.IsBasicLand,
		Dest:      dest,
		Limit:     1,
		Reveal:    true,
		Shuffle:   true,
		Reason:    "Caravan Vigil — a basic land card",
	}.Apply(ctx)
}
