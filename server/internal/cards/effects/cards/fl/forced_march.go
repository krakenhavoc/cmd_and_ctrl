package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Forced March — Sorcery {X}{B}{B}{B} (Mercadian Masques, S82 OLD, #5619):
//
//	"Destroy all creatures with mana value X or less."
//
// X is the value announced with the spell, and each creature's mana value is read
// as it resolves (CR 202.3): a token has mana value 0 and so always dies, a
// creature that is a copy has the copied mana value. All of them are destroyed as
// one batch.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "457dc141-bc72-42ad-b482-c78183462e34",
		Name:         "Forced March",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: func(g *game.Game, you uuid.UUID, c game.Card) bool {
				return effects.Creature()(g, you, c) && effects.ManaValueLE(ctx.X())(g, you, c)
			}}.Apply(ctx)
		},
	})
}
