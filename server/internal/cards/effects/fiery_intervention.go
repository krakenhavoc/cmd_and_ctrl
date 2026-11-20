package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fiery Intervention — Sorcery {4}{R} (Dominaria, #5530):
//
//	"Choose one —
//	 • Fiery Intervention deals 5 damage to target creature.
//	 • Destroy target artifact."
//
// Abrade's two modes, with 5 damage on the first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a8cd3a52-1056-4664-ae55-eb0ba3a43a15",
		Name:         "Fiery Intervention",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Fiery Intervention deals 5 damage to target creature.",
				TargetCreature("target creature")), ForTargets(DamageToTarget(0, 5))),
			Mode("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if ctx.HasMode(0) {
				return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 5}.Apply(ctx)
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
