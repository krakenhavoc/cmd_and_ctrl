package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Negate — "Counter target noncreature spell." Sandbox
// S20: the "noncreature" predicate is enforced — the picker only
// offers noncreature spells and the engine rejects a creature spell
// at announce (CR 601.2c). Previously the announce-time picker let the
// caster target any spell; the CounterTarget primitive does the
// rest. Social self-policing catches misplays in casual.
func init() {
	Register(Spec{
		OracleID: "3407fe41-fdd3-4119-8f70-4bc4590a379f",
		Name:     "Negate",
		Targets:  TargetSpell("target noncreature spell", Noncreature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
