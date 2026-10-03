package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Healing Grace — Instant {W}:
//
//	"Prevent the next 3 damage that would be dealt to any target this turn by a source of your choice. You gain 3 life."
//
// ADR 0108 §7 (#1904): CR 615.7's charged shield keyed on a source
// chosen as it resolves (CR 609.7a): the next 3 damage that source would
// deal to the target, across as many instances as it takes. A target gone
// by resolution gets no shield (CR 608.2b), and you still gain 3 life.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "e3b22777-4d5b-4f6c-a339-24d204bd007e",
		Name:         "Healing Grace",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := PreventDamageFromChosenSource(ShieldTheTarget).Charged(3).Apply(ctx); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
