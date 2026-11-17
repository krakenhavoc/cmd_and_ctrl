package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bathe in Light — Instant {1}{W}:
//
//	"Radiance — Choose a color. Target creature and each other creature
//	 that shares a color with it gain protection from the chosen color
//	 until end of turn."
//
// The creatures are read as the spell resolves (radianceSet, #5736) and
// the colour is chosen then too, by the caster, as the colour the
// creatures are protected FROM (ColorForProtection, #780). The two have
// nothing to do with each other (the card's ruling): the chosen colour
// need not be one the creatures share. Every creature that shares a
// colour is protected, the caster's own and everyone else's. A colourless
// target protects only itself. The set is carried to the colour's tail,
// so a table waiting on the answer is a restore point (#4253); nothing
// can change the board between the two.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8e862a8e-7954-43ad-9bc0-09ce7b9859f0",
		Name:         "Bathe in Light",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerProtect, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "protection"))},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ids := radianceSet(ctx, Creature())
			if len(ids) == 0 {
				return nil
			}
			then := ctx.ItemTail(radianceProtectionFromChosenColor)
			then.Carry = ids
			then.Words = []string{"Bathe in Light"}
			ChooseColorTail(game.ColorForProtection, ctx.Game, item.Controller, ctx.Source(), "Bathe in Light — choose a color", then)
			return nil
		},
	})
}
