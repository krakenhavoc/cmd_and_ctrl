package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Abrade — Instant {1}{R}:
//
//	"Choose one —
//	 • Abrade deals 3 damage to target creature.
//	 • Destroy target artifact."
//
// The reason red's two-mana removal slot belongs to Abrade rather
// than a straight burn spell: one card answers either half of the
// board. Both bullets target, which the S20 modal engine allows on a
// "choose one" (Register only rejects two targeted options when Max
// > 1), and the engine applies the CHOSEN option's target clause —
// so the picker offers creatures or artifacts depending on the mode,
// not the union of both.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f9db72dc-9a5b-48a4-a86e-7464d9a2166a",
		Name:         "Abrade",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: both modes destroy or damage a permanent
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Abrade deals 3 damage to target creature.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.DamageToTarget(0, 3))),
			effects.Mode("Destroy target artifact.",
				effects.TargetPermanent("target artifact", effects.Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: 3}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
