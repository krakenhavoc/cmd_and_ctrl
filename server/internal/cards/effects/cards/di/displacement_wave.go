package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Displacement Wave — Sorcery {X}{U}{U} (Magic Origins, S82 older sets):
//
//	"Return all nonland permanents with mana value X or less to their
//	 owners' hands."
//
// Evacuation over the nonland permanents whose mana value is at most the
// announced X, read as the spell resolves (CR 107.3, CR 202.3). One
// simultaneous bounce of every player's; a token cease to exist. A permanent
// with a mana value above X stays, so the sweep is partial (ADR 0126 §6). X
// = 0 still returns the tokens and the free permanents, so it is not an
// XMatters card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2b3fc609-7b19-441a-ab7e-58ccd70d955b",
		Name:         "Displacement Wave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.And(effects.Not(effects.Land()), effects.ManaValueLE(ctx.X()))}.Apply(ctx)
		},
	})
}
