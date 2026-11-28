package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Chaos Charm — Instant {R} (Mirage):
//
//	"Choose one —
//	 • Destroy target Wall.
//	 • Chaos Charm deals 1 damage to target creature.
//	 • Target creature gains haste until end of turn."
//
// Rith's Charm's shape: the bullet is chosen with the spell and only its own
// target is asked. A Wall is a permanent with the Wall subtype (a creature
// type), and each bullet checks its target again as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b81e5dbd-47a5-4994-91b7-e2a50c8e098d",
		Name:         "Chaos Charm",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Destroy target Wall.", effects.TargetPermanent("target Wall", effects.HasSubtype("Wall"))),
				game.Purpose{Answers: game.AnswerRemove}),
			effects.ModeWithPurpose(effects.Mode("Chaos Charm deals 1 damage to target creature.", effects.TargetCreature("target creature")),
				effects.ForTargets(effects.DamageToTarget(0, 1))),
			effects.ModeWithPurpose(effects.Mode("Target creature gains haste until end of turn.", effects.TargetCreature("target creature")),
				effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste"))),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			switch {
			case ctx.HasMode(0):
				return effects.DestroyTarget{Target: id}.Apply(ctx)
			case ctx.HasMode(1):
				return effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 1}.Apply(ctx)
			case ctx.HasMode(2):
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Chaos Charm — haste"}.Apply(ctx)
			}
			return nil
		},
	})
}
