package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sylvan Scavenging — Enchantment {1}{G}{G} (Foundations, #4465):
//
//	"At the beginning of your end step, choose one —
//	 • Put a +1/+1 counter on target creature you control.
//	 • Create a 3/3 green Raccoon creature token if you control a
//	   creature with power 4 or greater."
//
// A modal trigger (CR 700.2b): the mode, and the first bullet's target,
// are chosen as it goes on the stack. The second bullet's "if" is not an
// intervening "if" — it is part of the effect, so the bullet may be
// chosen with no big creature and is checked only as the ability
// resolves, against current power after the layers (Ferocious).
//
// No simplification.
func init() {
	t := AtYourEndStep("Sylvan Scavenging — choose one: a +1/+1 counter, or a Raccoon",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on target creature you control.",
			TargetCreature("target creature you control", YouControl()), PlusOneCounterOnTheModesTarget),
		ModeDoing("Create a 3/3 green Raccoon creature token if you control a creature with power 4 or greater.", nil,
			func(_ *game.StackItem, ctx *Context, _ int) error {
				if !b24ControlsCreatureWithPowerAtLeast(ctx.Game, ctx.Controller(), 4) {
					return nil
				}
				return CreateToken{Template: TokenCard("3/3 green Raccoon"), N: 1}.Apply(ctx)
			}),
	)
	Register(Spec{
		OracleID:     "398afdf2-4a12-430e-85c9-c68847230939",
		Name:         "Sylvan Scavenging",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
