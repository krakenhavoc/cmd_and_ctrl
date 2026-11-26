package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primal Growth — Sorcery {2}{G} (Commander 2015, S82 OLD #5546):
//
//	"Kicker—Sacrifice a creature. (You may sacrifice a creature in addition
//	 to any other costs as you cast this spell.)
//	 Search your library for a basic land card, put that card onto the
//	 battlefield, then shuffle. If this spell was kicked, instead search your
//	 library for up to two basic land cards, put them onto the battlefield,
//	 then shuffle."
//
// The sacrifice is the kicker (Final Flourish's), paid on casting; kicked, the
// search is for up to two basic lands INSTEAD of one. The lands enter
// untapped, and the library is shuffled once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "3cbdf37b-7fe3-4791-b33c-8591158b0ce5",
		Name:          "Primal Growth",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.KickerSacrifice("a creature", effects.Creature())},
		Purpose:       game.Purpose{Lands: 1, LandsUntapped: 1},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			n := 1
			if ctx.WasKicked() {
				n = 2
			}
			return effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneBattlefield,
				Limit:     n,
				Shuffle:   true,
				Reason:    "Primal Growth — search your library for basic land cards, put them onto the battlefield",
			}.Apply(ctx)
		},
	})
}
