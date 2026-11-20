package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Piercing Light — Instant {W} (Innistrad: Crimson Vow):
//
//	"Piercing Light deals 2 damage to target attacking or blocking
//	 creature. Scry 1."
//
// The target must be attacking or blocking when chosen and as the spell
// resolves (CR 608.2b). If it is no longer legal the whole spell is
// countered on resolution (all its targets are illegal), so no scry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2a6c7826-dcfe-4029-894e-6c82c4dd84b7",
		Name:         "Piercing Light",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target attacking or blocking creature", effects.AttackingOrBlocking()),
		Purpose: game.Purpose{
			Targets: game.ForTargets(effects.DamageToTarget(0, 2)),
			Answers: game.AnswerRemove,
			Scry:    1,
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return effects.Scry{N: 1}.Apply(ctx)
		},
	})
}
