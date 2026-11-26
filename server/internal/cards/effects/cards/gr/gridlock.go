package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gridlock — Instant {X}{U} (Gatecrash, S82 OLD #5565):
//
//	"Tap X target nonland permanents."
//
// Icy Blast's X-target clause over nonland permanents (the count is the
// announced X, CR 107.3), each target tapped as the spell resolves if it is
// still a legal target (CR 608.2b); an already-tapped permanent stays tapped.
// With X = 0 there are no targets and nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "049e9963-1b74-476f-9a81-dbfae543e1c5",
		Name:         "Gridlock",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.S82OldB411XTargetNonlandPermanents(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
