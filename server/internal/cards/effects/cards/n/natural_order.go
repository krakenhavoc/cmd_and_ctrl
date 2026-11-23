package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Order — Sorcery {2}{G}{G} (Mystery Booster 2, #5948):
//
//	"As an additional cost to cast this spell, sacrifice a green creature.
//	 Search your library for a green creature card, put it onto the
//	 battlefield, then shuffle."
//
// The sacrifice is paid while casting (CR 601.2f), so the creature is gone
// before the spell is on the stack; the search is not optional to find but
// may come up empty (CR 701.23b).
//
// No simplification.
func init() {
	green := effects.And(effects.Creature(), effects.OfColor("G"))
	effects.Register(effects.Spec{
		OracleID:       "8c1fe337-375a-4add-93b6-0ac39ed72b4f",
		Name:           "Natural Order",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerSacOutlet},
		AdditionalCost: effects.SacrificeCost("a green creature", green),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					return c.IsCreature() && c.HasColorWord("G")
				},
				Dest:    game.ZoneBattlefield,
				Limit:   1,
				Shuffle: true,
				Reason:  "Natural Order — a green creature card",
			}.Apply(ctx)
		},
	})
}
