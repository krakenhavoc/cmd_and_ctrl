package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Diabolic Revelation — Sorcery {X}{3}{B}{B} (Magic 2013):
//
//	"Search your library for up to X cards, put those cards into your hand, then
//	 shuffle."
//
// Harvest Season's X-limited search with no quality, to hand and not revealed.
// X of zero is still a search and still a shuffle: the search matches nothing
// rather than being skipped, because SearchLibrary reads a zero limit as one
// and the printed shuffle is unconditional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dcf438a7-adc4-4658-b75a-2dd633f595a7",
		Name:         "Diabolic Revelation",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Purpose:      game.Purpose{Tutors: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := ctx.X()
			match := func(game.Card) bool { return x > 0 }
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: match,
				Dest:      game.ZoneHand,
				Limit:     x,
				UpTo:      true,
				Shuffle:   true,
				Reason:    "Diabolic Revelation — up to X cards",
			}.Apply(ctx)
		},
	})
}
