package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trufflesnout — Creature — Boar {2}{G}, 2/2:
//
//	"When this creature enters, choose one —
//	 • Put a +1/+1 counter on this creature.
//	 • You gain 4 life."
//
// A modal enters trigger (CR 700.2): the bullet is chosen as the
// trigger is put on the stack. If Trufflesnout has left by the time the
// counter bullet resolves, nothing gets the counter (CR 400.7).
//
// No simplification.
func init() {
	enters := WhenThisEnters("Trufflesnout — choose one", func(*game.Game, *game.StackItem) error { return nil })
	enters.Modes = ChooseOne(
		ModeDoing("Put a +1/+1 counter on this creature.", nil, putAPlusOneCounterOnThisCreatureMode),
		ModeDoing("You gain 4 life.", nil, func(item *game.StackItem, ctx *Context, _ int) error {
			return GainLife{Player: item.Controller, Amount: 4}.Apply(ctx)
		}),
	)
	Register(Spec{
		OracleID:     "8a15841a-dcb6-454a-b865-43a067f42b4e",
		Name:         "Trufflesnout",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
