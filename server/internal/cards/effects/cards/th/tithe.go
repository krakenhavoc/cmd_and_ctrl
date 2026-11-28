package cardsth

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tithe — Instant {W} (Visions, S82 OLD batch 787, #5631):
//
//	"Search your library for a Plains card. If target opponent controls more
//	 lands than you, you may search your library for an additional Plains
//	 card. Reveal those cards, put them into your hand, then shuffle."
//
// One search for Plains cards (the subtype, not only basic ones), up to one,
// or up to two when the target opponent controls strictly more lands than you
// as the spell resolves (a search may find fewer than it could, CR 701.23b);
// the cards are revealed, go to hand, and the library is shuffled. The
// opponent is the spell's only target, so with it gone the spell does not
// resolve and nothing is searched (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "318304f5-0929-4cd3-97e6-08ac1ac48aa6",
		Name:         "Tithe",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Purpose:      game.Purpose{Tutors: 2},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			g := ctx.Game
			you := ctx.Controller()
			limit := 1
			if opp := effects.TargetedPlayerOf(ctx); opp != uuid.Nil && effects.B11OpponentControlsMoreThanYou(g, you, opp, game.Card.IsLand) {
				limit = 2
			}
			return effects.SearchLibrary{
				Player: you,
				Predicate: func(c game.Card) bool {
					return c.HasSubtypeWord("Plains")
				},
				Dest:    game.ZoneHand,
				Limit:   limit,
				Reveal:  true,
				Shuffle: true,
				Reason:  "Tithe — a Plains card",
			}.Apply(effects.NewContext(g, item))
		},
	})
}
