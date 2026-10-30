package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Spot's Portal — {2}{B} Instant (Marvel Super Heroes Commander,
// #4455):
//
//	"Put target creature on the bottom of its owner's library. You lose 2
//	 life unless you control a Villain."
//
// The Villain test is made as the spell resolves, after the creature has
// left (CR 608.2c), so a Villain that was the target does not count. If the
// target is gone the spell does nothing at all: it is countered on
// resolution (CR 608.2b) and the life is not lost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ceac6ce-acde-42ab-b79d-ff153074d2d8",
		Name:         "The Spot's Portal",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
				return nil
			}
			if err := ctx.Game.TuckToLibraryForEffect(t.ID, true); err != nil {
				return err
			}
			if b05ControlsSubtype(ctx.Game, item.Controller, ctx.TextSubtype("Villain")) {
				return nil
			}
			return GainLife{Player: item.Controller, Amount: -2}.Apply(ctx)
		},
	})
}
