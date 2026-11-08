package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Succumb to the Cold — Instant {2}{U} (Wilds of Eldraine, #4471):
//
//	"Tap one or two target creatures an opponent controls. Put a stun
//	 counter on each of them. (If a permanent with a stun counter would
//	 become untapped, remove one from it instead.)"
//
// One target clause of one or two creatures an opponent controls (Amazing
// Acrobatics' count). Each target still legal as the spell resolves is
// tapped and gets a stun counter (CR 608.2b); the stun counter is the
// engine's (CR 122.1d), which replaces the next untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "73992788-8084-459c-8375-9ccc8579c5d7",
		Name:         "Succumb to the Cold",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("one or two target creatures an opponent controls", OpponentControls()).WithCount(1, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if err := (TapTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterStun, N: 1}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
