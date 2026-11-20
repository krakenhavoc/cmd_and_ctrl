package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Start from Scratch — Sorcery — Lesson {2}{R} (Strixhaven: School of
// Mages, #5502):
//
//	"Choose one —
//	 • Start from Scratch deals 1 damage to any target.
//	 • Destroy target artifact."
//
// Abrade's two bullets with "any target" (a creature, a player, a
// planeswalker or a battle) for the damage. Each bullet carries its own
// target clause, judged at announce and again as the spell resolves.
// Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8e255ec-7d66-43f6-b5f0-712f5957687f",
		Name:         "Start from Scratch",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Start from Scratch deals 1 damage to any target.", TargetAny()), ForTargets(DamageToTarget(0, 1))),
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			target := item.Targets[0]
			switch {
			case ctx.HasMode(0):
				return DealDamage{Source: ctx.Source(), Target: target.ID, Amount: 1}.Apply(ctx)
			case ctx.HasMode(1):
				if target.Kind != game.TargetCard {
					return nil
				}
				return DestroyTarget{Target: target.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
