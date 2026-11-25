package cardsri

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// River Heralds' Boon — Instant {1}{G} (Ixalan, #5532):
//
//	"Put a +1/+1 counter on target creature and a +1/+1 counter on up to
//	 one target Merfolk."
//
// Two clauses chosen as the spell is cast: any creature, and up to one
// Merfolk (any controller's), which may be the same creature (CR 115.3
// only bars one target being chosen twice for the same word). Each clause
// that is still legal as the spell resolves gets its counter (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "682eb224-cab4-4829-810b-72e63a9ceb5d",
		Name:         "River Heralds' Boon",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetCreature("up to one target Merfolk", effects.Subtype("Merfolk")).WithCount(0, 1),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for slot := 0; slot < 2; slot++ {
				if t, ok := ctx.ClauseTarget(slot); ok {
					ids = append(ids, t.ID)
				}
			}
			for _, id := range ids {
				if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
