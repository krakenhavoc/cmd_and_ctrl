package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Access Denied — Instant {3}{U}{U} (Kamigawa: Neon Dynasty Commander, S82
// #5493):
//
//	"Counter target spell. Create X 1/1 colorless Thopter artifact
//	 creature tokens with flying, where X is that spell's mana value."
//
// The mana value is read off the spell before the counter takes it off the
// stack (CR 608.2h: information about it as it last existed on the stack),
// with {X} counted as the value chosen (CR 202.3e). A spell that can't be
// countered is still "that spell", so the Thopters are made regardless; if
// the target is no longer legal, Access Denied does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "353dd446-2ea6-41c9-a10f-5a03e4de19b4",
		Name:         "Access Denied",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			targets := ctx.LegalTargets()
			if len(targets) == 0 {
				return nil
			}
			x := effects.SpellManaValueForEffect(ctx.Game, targets[0].ID)
			if err := (effects.CounterTarget{StackID: targets[0].ID}).Apply(ctx); err != nil {
				return err
			}
			return effects.CreateToken{Template: effects.TokenCard("1/1 colorless Thopter artifact with flying"), N: x}.Apply(ctx)
		},
	})
}
