package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Go Ninja Go — {R}{W} Sorcery (Teenage Mutant Ninja Turtles, #4457):
//
//	"Choose one or both —
//	 • Exile target creature you control, then return it to the
//	   battlefield under its owner's control.
//	 • Go Ninja Go deals damage equal to the greatest power among
//	   creatures you control to target creature an opponent controls."
//
// Amazing Acrobatics' "choose one or both" (CR 700.2), each bullet a
// ModeDoing reading its own target group, run in printed order
// (CR 608.2c). The first bullet is a Flicker with no Controller, so the
// card comes back under its owner's control; the engine holds the second
// bullet until the flicker has finished. The second reads "the greatest
// power among creatures you control" as it resolves, after the flicker, so
// a creature that came back counts with its new power. With no creature,
// the amount is 0 and nothing is dealt.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "46f77c8c-6817-4aea-a80a-d91f508ccbd0",
		Name:         "Go Ninja Go",
		Completeness: CompletenessFull,
		Modes: ChooseN("Choose one or both", 1, 2,
			ModeDoing("Exile target creature you control, then return it to the battlefield under its owner's control.",
				TargetCreature("target creature you control", YouControl()),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return Flicker{Target: t.ID}.Apply(ctx)
				}),
			ModeDoing("Go Ninja Go deals damage equal to the greatest power among creatures you control to target creature an opponent controls.",
				TargetCreature("target creature an opponent controls", OpponentControls()),
				func(item *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					n := b42GreatestPowerControlledBy(ctx.Game, ctx.Controller())
					return DealDamage{Source: item.SourceCardID, Target: t.ID, Amount: n}.Apply(ctx)
				}),
		),
	})
}
