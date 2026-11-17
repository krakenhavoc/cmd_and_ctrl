package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ready to Rumble — Sorcery {4}{R} (Streets of New Capenna, #5491):
//
//	"Choose one —
//	 • Ready to Rumble deals 5 damage to target creature or planeswalker.
//	 • Destroy target artifact."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04dd9395-847b-46f7-a669-9b0df89d01ff",
		Name:         "Ready to Rumble",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Ready to Rumble deals 5 damage to target creature or planeswalker.",
				TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker()))), ForTargets(DamageToTarget(0, 5))),
			Mode("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return DealDamage{Source: ctx.Source(), Target: target, Amount: 5}.Apply(ctx)
			case ctx.HasMode(1):
				return DestroyTarget{Target: target}.Apply(ctx)
			}
			return nil
		},
	})
}
