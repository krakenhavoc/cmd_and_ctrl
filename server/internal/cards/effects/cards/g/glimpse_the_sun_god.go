package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Glimpse the Sun God — Instant {X}{W} (Born of the Gods, S82 OLD b377, #5559):
//
//	"Tap X target creatures. Scry 1."
//
// Icy Blast's X targets (OldB347XTargetCreatures), the tap without a freeze,
// then the scry (a fixed rider, declared in Purpose).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "87cc3a67-b4f0-41f0-9e34-745ef23aa479",
		Name:         "Glimpse the Sun God",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 1},
		Targets:      effects.OldB347XTargetCreatures(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			for _, id := range ids {
				if err := (effects.TapTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.Scry{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
