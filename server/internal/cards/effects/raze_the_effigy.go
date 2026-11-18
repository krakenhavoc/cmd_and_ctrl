package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raze the Effigy — Instant {R} (Innistrad: Midnight Hunt, #5497):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Target attacking creature gets +2/+2 until end of turn."
//
// Abrade's modal shape; each bullet carries its own target clause.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9286ddc5-8151-4129-aaf6-2a2abe7047e5",
		Name:         "Raze the Effigy",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Destroy target artifact.",
				TargetPermanent("target artifact", Artifact())), game.Purpose{Answers: game.AnswerRemove}),
			ModeWithPurpose(Mode("Target attacking creature gets +2/+2 until end of turn.",
				TargetCreature("target attacking creature", AttackingCreature())), ForTargets(BenefitUntilEOT(0, 2, 2))),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			target := item.Targets[0].ID
			switch {
			case ctx.HasMode(0):
				return DestroyTarget{Target: target}.Apply(ctx)
			case ctx.HasMode(1):
				return BoostUntilEOT{Target: target, Power: 2, Toughness: 2, Label: "Raze the Effigy — +2/+2"}.Apply(ctx)
			}
			return nil
		},
	})
}
