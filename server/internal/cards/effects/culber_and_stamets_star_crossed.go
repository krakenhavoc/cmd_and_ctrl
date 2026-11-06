package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Culber and Stamets, Star-Crossed — Legendary Creature — Human Doctor
// Artificer {3}{G}, 3/4 (Star Trek, #4973):
//
//	"When Culber and Stamets enter, choose one —
//	 • Put a +1/+1 counter on target creature.
//	 • Add two mana of any one color."
//
// A modal entry trigger (Charming Prince's shape): the mode is chosen as
// the trigger goes on the stack (CR 700.2a) and the first mode's target
// with it. The second mode is not a mana ability (it is a triggered
// ability, so it uses the stack, CR 605.1a): the two mana of one colour
// the controller picks go to their pool as it resolves and empty with the
// pool at the end of the step.
//
// No simplification.
func init() {
	etb := WhenThisEnters("Culber and Stamets, Star-Crossed — choose one",
		func(*game.Game, *game.StackItem) error { return nil })
	etb.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on target creature.", TargetCreature("target creature"),
			PlusOneCountersOnTheModesTarget(1)),
		ModeDoing("Add two mana of any one color.", nil, func(_ *game.StackItem, ctx *Context, _ int) error {
			return AddMana{Produced: OneColorOfAmount(2)}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "5b8e8ef2-2f04-4485-ae12-08f52dac81a2",
		Name:         "Culber and Stamets, Star-Crossed",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{etb},
	})
}
