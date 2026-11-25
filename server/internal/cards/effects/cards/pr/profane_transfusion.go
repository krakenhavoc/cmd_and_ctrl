package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Profane Transfusion — Sorcery {6}{B}{B}{B} (Commander Legends):
//
//	"Two target players exchange life totals. You create an X/X colorless
//	 Horror artifact creature token, where X is the difference between those
//	 players' life totals."
//
// Axis of Mortality's exchange (both totals read before either is written, each
// half through the CR 614 window), and the token is made from a continuation
// that reads the two totals once the exchange has settled, so a replacement
// that changed what moved is honoured. An exchange of equal totals still makes
// the token, as a 0/0 that dies (CR 704.5f). With one target gone by resolution
// nothing is exchanged and, with no pair of players to compare, no token is made.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2eaee808-b72a-4950-ab4f-ab8e8d12f838",
		Name:         "Profane Transfusion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("two target players").WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			legal := ctx.LegalTargets()
			if len(legal) != 2 || legal[0].Kind != game.TargetPlayer || legal[1].Kind != game.TargetPlayer {
				return nil
			}
			return effects.ExchangeLifeTotalsThen(ctx, legal[0].ID, legal[1].ID, ctx.ItemTail(profaneTransfusionToken))
		},
	})
}

// profaneTransfusionToken is "you create an X/X Horror, where X is the
// difference between those players' life totals", read after the exchange. An
// on-disk key: never renamed, never reused.
var profaneTransfusionToken = effects.CardTail("profane-transfusion/token", func(ctx *effects.Context, _ game.TailOutcome) error {
	legal := ctx.LegalTargets()
	if len(legal) != 2 {
		return nil
	}
	a, b := ctx.PlayerByID(legal[0].ID), ctx.PlayerByID(legal[1].ID)
	if a == nil || b == nil {
		return nil
	}
	x := a.Life - b.Life
	if x < 0 {
		x = -x
	}
	return effects.S82OldB083HorrorToken(ctx, x)
})
