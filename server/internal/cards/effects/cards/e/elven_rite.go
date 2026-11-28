package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Rite — Sorcery {1}{G} (Stronghold, S82 OLD b749, #5627):
//
//	"Distribute two +1/+1 counters among one or two target creatures."
//
// Abzan Charm's distribution (CR 601.2d: each target gets at least one, the
// split is announced with the targets), as a spell of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "755dbf29-7484-444f-b5df-29ec4b4bbc82",
		Name:         "Elven Rite",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("one or two target creatures").WithCount(1, 2).Dividing(effects.Divide(2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PutDividedCounters(ctx, game.CounterPlusOne)
		},
	})
}
