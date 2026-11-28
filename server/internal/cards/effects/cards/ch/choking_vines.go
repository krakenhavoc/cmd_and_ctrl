package cardsch

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Choking Vines — {X}{G} Instant (Weatherlight, S82 OLD batch 771, #5629):
//
//	"Cast this spell only during the declare blockers step.
//	 X target attacking creatures become blocked. Choking Vines deals 1 damage to
//	 each of those creatures. (This spell works on creatures that can't be
//	 blocked.)"
//
// Fog Patch's cast window and become-blocked effect (Game.BecomeBlockedForEffect,
// CR 509.1h) over Immoral Bargain's X targets, attacking creatures only. A target
// that is no longer an attacking creature as the spell resolves is skipped (CR
// 608.2b). The attackers become blocked first, then each is dealt 1 damage as one
// event.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b6b565e-c0b5-48f9-b3e0-4d5d4b065402",
		Name:         "Choking Vines",
		Completeness: effects.CompletenessFull,
		CastCondition: func(g *game.Game, _ uuid.UUID, _ game.Card) bool {
			return g.Turn.Step == game.StepDeclareBlockers
		},
		CastConditionLabel: "Cast this spell only during the declare blockers step.",
		Targets:            effects.VsXTarget("X target attacking creatures", effects.AttackingCreature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			var ids []uuid.UUID
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetCard {
					ids = append(ids, t.ID)
				}
			}
			if len(ids) == 0 {
				return nil
			}
			ctx.Game.BecomeBlockedForEffect(ids)
			return ctx.Game.DamageInstanceForEffect(func() error {
				for _, id := range ids {
					if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
