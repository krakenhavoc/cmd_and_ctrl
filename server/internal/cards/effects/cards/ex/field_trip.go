package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Field Trip — Sorcery {2}{G} (EDHREC rank 18535):
//
//	"Search your library for a basic Forest card, put that card onto
//	 the battlefield tapped, then shuffle.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// The search asks for a pick, so learn waits for it as a later step
// (CR 608.2c); a player is not asked what to discard before the Forest
// is chosen and the library shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29423b70-7cb7-4617-898c-78a38eb233ad",
		Name:         "Field Trip",
		Purpose:      game.Purpose{Lands: 1, Learn: 1},
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     effects.IsBasicLandWithSubtype("forest"),
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Field Trip — a basic Forest card",
				}.Apply(ctx)
			}),
			effects.Learn{},
		),
	})
}
