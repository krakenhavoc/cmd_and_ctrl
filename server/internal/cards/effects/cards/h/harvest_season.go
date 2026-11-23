package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Harvest Season — Sorcery {2}{G} (Starter Commander Decks):
//
//	"Search your library for up to X basic land cards, where X is the number
//	 of tapped creatures you control. Put those cards onto the battlefield
//	 tapped, then shuffle."
//
// Traverse the Outlands' search, with X counted as the spell resolves over
// the tapped creatures you control. X of zero is still a search and still a
// shuffle: the search matches nothing rather than being skipped, because
// SearchLibrary reads a zero limit as one and the printed shuffle is
// unconditional.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec6d52f9-1c8f-48da-951c-be8d7813c7be",
		Name:         "Harvest Season",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := len(effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.And(effects.Creature(), effects.TappedPermanent())))
			match := effects.IsBasicLand
			if x <= 0 {
				match = func(game.Card) bool { return false }
			}
			return effects.SearchLibrary{
				Player:        ctx.Controller(),
				Predicate:     match,
				Dest:          game.ZoneBattlefield,
				Limit:         x,
				Shuffle:       true,
				TappedOnEntry: true,
				Reason:        "Harvest Season — up to X basic lands, tapped",
			}.Apply(ctx)
		},
	})
}
