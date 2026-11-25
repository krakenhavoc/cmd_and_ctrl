package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reduce // Rubble — Instant // Sorcery, Aftermath (Amonkhet, S82 OLD batch 243, #5535):
//
//	Reduce {2}{U}: "Counter target spell unless its controller pays {3}."
//	Rubble {2}{R}: "Aftermath. Up to three target lands don't untap during
//	 their controller's next untap step."
//
// Two registered faces. Reduce is Complicate's soft counter for {3}. Rubble
// takes up to three target lands, each held through its own controller's next
// untap step (CR 502.3); it is cast only from a graveyard and exiled as it
// resolves (CR 702.127a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cac993d5-28a1-48b6-9c2f-f216cb5f521d",
		Name:         "Reduce",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.CounterTheTargetSpellUnlessPaid("Reduce", "{3}")(ctx.Game, item)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     "cac993d5-28a1-48b6-9c2f-f216cb5f521d#1",
		Name:         "Rubble",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("up to three target lands", effects.Land()).WithCount(0, 3),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			return effects.DoesntUntapNextUntapStep{Targets: ids, Label: "Rubble — doesn't untap during its controller's next untap step"}.Apply(ctx)
		},
	})
}
