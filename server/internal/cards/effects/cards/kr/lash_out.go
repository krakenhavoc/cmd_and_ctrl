package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lash Out — Instant {1}{R} (Commander 2011, #5574):
//
//	"Lash Out deals 3 damage to target creature. Clash with an opponent. If
//	 you win, Lash Out deals 3 damage to that creature's controller."
//
// The creature's controller is read before the damage, so a creature the
// 3 damage kills still has its last controller hit when the clash is won
// (CR 608.2h). A target that became illegal means the whole spell does
// nothing, clash included (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "398d57bb-4653-4652-aa42-7cea4317be0a",
		Name:         "Lash Out",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 3)), Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			controller, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			then := ctx.Tail(lashOutClashed)
			then.Carry = []uuid.UUID{controller}
			return effects.ClashWithAnOpponent{Then: then}.Apply(ctx)
		},
	})
}

// lashOutClashed is "If you win, Lash Out deals 3 damage to that creature's
// controller" (the tail's Carry). An on-disk key.
var lashOutClashed = effects.CardTail("then/lash-out-clashed", func(ctx *effects.Context, o game.TailOutcome) error {
	if !effects.WonTheClash(o) || len(o.Tail.Carry) == 0 {
		return nil
	}
	return effects.DealDamage{Source: ctx.Source(), Target: o.Tail.Carry[0], Amount: 3}.Apply(ctx)
})
