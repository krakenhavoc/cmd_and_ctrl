package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Negate — "Counter target noncreature spell." Sandbox
// simplification: S14 does not enforce the "noncreature" predicate
// (S20 smart-cast UI will). The announce-time picker lets the
// caster target any spell; the CounterTarget primitive does the
// rest. Social self-policing catches misplays in casual.
func init() {
	Register(Spec{
		OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f",
		Name:     "Negate",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
