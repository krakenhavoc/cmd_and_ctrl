package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2a6c7826-dcfe-4029-894e-6c82c4dd84b7",
		Name:         "Piercing Light",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking or blocking creature", AttackingOrBlocking()),
		Purpose: game.Purpose{
			Targets: game.ForTargets(DamageToTarget(0, 2)),
			Answers: game.AnswerRemove,
			Scry:    1,
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (DealDamage{Source: item.SourceCardID, Target: id, Amount: 2}).Apply(ctx); err != nil {
				return err
			}
			return Scry{N: 1}.Apply(ctx)
		},
	})
}
