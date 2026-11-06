package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Might of the Meek — Instant {R} (Bloomburrow, #4467):
//
//	"Target creature gains trample until end of turn. It also gets +1/+0
//	 until end of turn if you control a Mouse.
//	 Draw a card."
//
// "If you control a Mouse" is read as the spell resolves (CR 608.2c), on
// the battlefield through the layers and through the spell's own text
// (ctx.TextSubtype, CR 612.2), so a Mouse that left in response no longer
// counts. The draw is its own step, after the pump. A target that left in
// response counters the spell for want of a target (CR 608.2b), and then
// nothing is drawn either.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19b5c94b-b4fa-41a9-b649-51c4e6f03602",
		Name:         "Might of the Meek",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      game.Purpose{Draws: 1, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "trample"))},
		OnResolve:    ResolveSteps(Instruction(mightOfTheMeekPump), DrawCards{N: 1}),
	})
}

// mightOfTheMeekPump is the first sentence: trample, and +1/+0 while you
// control a Mouse.
func mightOfTheMeekPump(ctx *Context) error {
	t, ok := ctx.ClauseTarget(0)
	if !ok || t.Kind != game.TargetCard {
		return nil
	}
	if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Might of the Meek — trample"}).Apply(ctx); err != nil {
		return err
	}
	if !ControlsA(ctx.TextSubtype("Mouse"))(ctx.Game, ctx.Controller()) {
		return nil
	}
	return BoostUntilEOT{Target: t.ID, Power: 1, Label: "Might of the Meek — +1/+0"}.Apply(ctx)
}
