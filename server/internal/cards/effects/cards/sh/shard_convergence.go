package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shard Convergence — Sorcery {3}{G} (Conflux):
//
//	"Search your library for a Plains card, an Island card, a Swamp card, and
//	 a Mountain card. Reveal those cards, put them into your hand, then
//	 shuffle."
//
// Gem of Becoming's one search for up to one card of each subtype with a set
// rule that each pick fills a different slot, widened to four: a land that is
// both a Plains and an Island fills only one of them (CR 701.23d lets a
// library lack one, so fewer may be found), and the shuffle happens either
// way.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c51fc4f2-76a6-4337-ab8f-1b3e84f82019",
		Name:         "Shard Convergence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue, Tutors: 4},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SearchLibrary{
				Player: ctx.Controller(),
				Predicate: func(c game.Card) bool {
					for _, t := range effects.OldB506ShardTypes {
						if c.HasSubtypeWord(t) {
							return true
						}
					}
					return false
				},
				Dest:    game.ZoneHand,
				Limit:   len(effects.OldB506ShardTypes),
				UpTo:    true,
				Reveal:  true,
				Shuffle: true,
				Rule:    effects.OldB506OneOfEachShard,
				Reason:  "Shard Convergence — a Plains card, an Island card, a Swamp card and a Mountain card",
			}.Apply(ctx)
		},
	})
}
