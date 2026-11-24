package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dramatic Entrance — Instant {3}{G}{G} (Secret Lair Drop):
//
//	"You may put a green creature card from your hand onto the battlefield."
//
// Norwood Priestess's put (PutFromHandOntoBattlefield over a green creature card,
// optional), as an instant's whole effect. A hand with no green creature card
// asks nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dda8bf18-9396-45e6-829c-8691f49f7290",
		Name:         "Dramatic Entrance",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PutFromHandOntoBattlefield{
				Match:    effects.And(effects.Creature(), effects.OfColor("G")),
				Optional: true,
				Label:    "Dramatic Entrance — you may put a green creature card from your hand onto the battlefield",
			}.Apply(ctx)
		},
	})
}
