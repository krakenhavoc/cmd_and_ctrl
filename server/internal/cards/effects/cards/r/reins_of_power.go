package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reins of Power — Instant {2}{U}{U}:
//
//	"Untap all creatures you control and all creatures target
//	 opponent controls. You and that opponent each gain control of
//	 all creatures the other controls until end of turn. Those
//	 creatures gain haste until end of turn."
//
// A two-way Insurrection: both snapshots (the caster's creatures and
// the chosen opponent's) are taken from who controls what BEFORE
// anything moves, so the untap and the swap both read the original
// board rather than one another's output. The swap itself is two runs
// of GainControl — the caster's creatures go to the opponent, the
// opponent's go to the caster — because ExchangeControl (control.go)
// is CR 701.12's ALL-OR-NOTHING single-pair primitive and this clause
// exchanges two disjoint SETS, whatever size each is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96381388-1e82-411c-8290-1f3e909d3b5f",
		Name:         "Reins of Power",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.SwapEachPlayersCreatures(ctx, "Reins of Power")
		},
	})
}
