package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Venom's Hunger — Sorcery {4}{B} (Marvel's Spider-Man, #4460):
//
//	"This spell costs {2} less to cast if you control a Villain.
//	 Destroy target creature. You gain 2 life."
//
// The discount is a plain self cost modifier over the caster's own
// battlefield (the spell is on the stack while the price is worked out, so it
// never counts itself). The life is gained on resolution whether or not the
// creature was destroyed (indestructible); an illegal target fizzles the whole
// spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8388faa0-f64e-4b01-836e-617f99b808a4",
		Name:         "Venom's Hunger",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if you control a Villain.",
				YouControlA(HasSubtype("Villain"))),
		},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			if err := (DestroyTarget{Target: item.Targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 2}.Apply(ctx)
		},
	})
}
