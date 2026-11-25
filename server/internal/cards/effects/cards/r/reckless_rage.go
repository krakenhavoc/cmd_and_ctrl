package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reckless Rage — Instant {R} (Rivals of Ixalan, S82 OLD batch 217, #5531):
//
//	"Reckless Rage deals 4 damage to target creature you don't control and 2
//	 damage to target creature you control."
//
// Two target clauses (CR 115.1). Both amounts are dealt at the same time, in
// one damage event, to the targets still legal as the spell resolves
// (CR 608.2b): an illegal one takes nothing and the other still does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d7e2b2ba-cb5e-41f1-ad33-baf56c0b72ed",
		Name:         "Reckless Rage",
		Completeness: effects.CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Targets: game.ForTargets(effects.DamageToTarget(0, 4), effects.DamageToTarget(1, 2)),
		},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you don't control", effects.OpponentControls()),
			effects.TargetCreature("target creature you control", effects.YouControl()),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return ctx.Game.DamageInstanceForEffect(func() error {
				for slot, amount := range []int{4, 2} {
					t, ok := ctx.ClauseTarget(slot)
					if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
						continue
					}
					if err := (effects.DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: amount}).Apply(ctx.AsGroupMember()); err != nil {
						return err
					}
				}
				return nil
			})
		},
	})
}
