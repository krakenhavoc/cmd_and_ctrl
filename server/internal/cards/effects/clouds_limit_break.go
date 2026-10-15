package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:     "93df61b5-6abe-4603-9a6e-cba90a834173",
		Name:         "Cloud's Limit Break",
		Completeness: CompletenessFull,
		Modes: Tiered(
			SpreeModeDoing("Cross-Slash — Destroy target tapped creature.", "{0}",
				TargetCreature("target tapped creature", tappedPermanent()),
				DestroyTheModesTarget),
			SpreeModeDoing("Blade Beam — Destroy any number of target tapped creatures with different controllers.", "{1}",
				TargetCreature("any number of target tapped creatures with different controllers", tappedPermanent()).
					WithCount(0, 0).EachDifferent(EachDifferentController()),
				destroyEachOfTheModesTargets),
			ModeWithPurpose(SpreeModeDoing("Omnislash — Destroy all tapped creatures.", "{3}{W}", nil,
				func(_ *game.StackItem, ctx *Context, _ int) error {
					return DestroyAllMatching{Match: And(Creature(), tappedPermanent())}.Apply(ctx)
				}), game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}}),
		),
	})
}

// destroyEachOfTheModesTargets is "destroy any number of target …" as a
// modal bullet's body: every still-legal target of this occurrence
// (CR 608.2b), destroyed simultaneously.
func destroyEachOfTheModesTargets(_ *game.StackItem, ctx *Context, occ int) error {
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
