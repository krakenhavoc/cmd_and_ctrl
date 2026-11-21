package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grim Affliction — Instant {2}{B} (New Phyrexia, #5575):
//
//	"Put a -1/-1 counter on target creature, then proliferate. (Choose any number of permanents and/or players, then give each another counter of each kind already there.)"
//
// One -1/-1 counter on the target creature, then a proliferate its
// controller chooses (CR 701.34a). The proliferate is the last
// instruction, so nothing waits behind its prompt.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "411ba1a2-30bb-4618-b874-6be0d996f07d",
		Name:         "Grim Affliction",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if err := (effects.AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
