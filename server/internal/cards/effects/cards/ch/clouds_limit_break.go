package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cloud's Limit Break — Instant {1}{W} (#3656, #3777):
//
//	"Tiered (Choose one additional cost.)
//	 • Cross-Slash — {0} — Destroy target tapped creature.
//	 • Blade Beam — {1} — Destroy any number of target tapped creatures
//	   with different controllers.
//	 • Omnislash — {3}{W} — Destroy all tapped creatures."
//
// Tiered proof card (CR 702.183a): exactly one bullet, its cost added
// to the printed {1}{W}. Blade Beam's "different controllers" is the
// EachDifferentController set rule the announce gate enforces, and its
// creatures are destroyed together, as one event. Omnislash reads the
// tapped creatures as it resolves (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93df61b5-6abe-4603-9a6e-cba90a834173",
		Name:         "Cloud's Limit Break",
		Completeness: effects.CompletenessFull,
		Modes: effects.Tiered(
			effects.SpreeModeDoing("Cross-Slash — Destroy target tapped creature.", "{0}",
				effects.TargetCreature("target tapped creature", effects.TappedPermanent()),
				effects.DestroyTheModesTarget),
			effects.SpreeModeDoing("Blade Beam — Destroy any number of target tapped creatures with different controllers.", "{1}",
				effects.TargetCreature("any number of target tapped creatures with different controllers", effects.TappedPermanent()).
					WithCount(0, 0).EachDifferent(effects.EachDifferentController()),
				destroyEachOfTheModesTargets),
			effects.ModeWithPurpose(effects.SpreeModeDoing("Omnislash — Destroy all tapped creatures.", "{3}{W}", nil,
				func(_ *game.StackItem, ctx *effects.Context, _ int) error {
					return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.TappedPermanent())}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		),
	})
}

// destroyEachOfTheModesTargets is "destroy any number of target …" as a
// modal bullet's body: every still-legal target of this occurrence
// (CR 608.2b), destroyed simultaneously.
func destroyEachOfTheModesTargets(_ *game.StackItem, ctx *effects.Context, occ int) error {
	var ids []uuid.UUID
	for _, t := range ctx.ModeTargets(occ) {
		if t.Kind == game.TargetCard && ctx.IsTargetLegal(t) {
			ids = append(ids, t.ID)
		}
	}
	if len(ids) > 0 {
		ctx.Game.DestroyPermanentsForEffect(ids)
	}
	return nil
}
