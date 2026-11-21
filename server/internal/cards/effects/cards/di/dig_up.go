package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dig Up — Sorcery {G}:
//
//	"Cleave {1}{B}{B}{G} (You may cast this spell for its cleave cost.
//	 If you do, remove the words in square brackets.)
//	 Search your library for a [basic land] card, [reveal it,] put it
//	 into your hand, then shuffle."
//
// Hard-cast it fetches a basic land and reveals it; cleaved it fetches
// any card and reveals nothing (CR 702.148a). The brackets touch no
// target clause (there is none), hence Cleave's nil.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ae96fe7e-8d89-4f44-9cc4-3c9b124fdc4b",
		Name:             "Dig Up",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Tutors: 1},
		AlternativeCosts: []game.AlternativeCost{effects.Cleave("{1}{B}{B}{G}", nil)},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			search := effects.SearchLibrary{
				Player:    ctx.Controller(),
				Predicate: effects.IsBasicLand,
				Dest:      game.ZoneHand,
				Limit:     1,
				Reveal:    true,
				Shuffle:   true,
				Reason:    "Dig Up — a basic land card",
			}
			if ctx.PaidAltCost("cleave") {
				search.Predicate, search.Reveal = nil, false
				search.Reason = "Dig Up — a card"
			}
			return search.Apply(ctx)
		},
	})
}
