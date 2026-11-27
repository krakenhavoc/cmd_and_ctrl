package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gaea's Balance — Sorcery {3}{G} (Apocalypse, S82 OLD, #5614):
//
//	"As an additional cost to cast this spell, sacrifice five lands.
//	 Search your library for a land card of each basic land type, put those
//	 cards onto the battlefield, then shuffle."
//
// The five lands are sacrificed as the additional cost, so they are gone
// before the search (CR 601.2h). Shard Convergence's one-card-per-slot search,
// widened to the five basic land types: a land with two types (a dual) fills
// only one slot, a library may lack a type (CR 701.23d) and the shuffle
// happens either way. The cards arrive untapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "5aadbd03-c459-4893-aa1b-efcf193faa1b",
		Name:           "Gaea's Balance",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.SacrificeNCost(5, "five lands", effects.Land()),
		Purpose:        game.Purpose{Answers: game.AnswerValue, Tutors: 5},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					for _, t := range effects.OldB629BasicLandTypes {
						if c.HasSubtypeWord(t) {
							return true
						}
					}
					return false
				},
				Dest:    game.ZoneBattlefield,
				Limit:   len(effects.OldB629BasicLandTypes),
				UpTo:    true,
				Shuffle: true,
				Rule:    effects.OldB629OneLandOfEachBasicType,
				Reason:  "Gaea's Balance — a land card of each basic land type",
			}.Apply(ctx)
		},
	})
}
