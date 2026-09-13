package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dispel — Instant, {U} (EDHREC rank 926):
//
//	"Counter target instant spell."
//
// Negate's narrower cousin: the clause admits instants only, so a
// sorcery, a creature or an ability is refused at announce.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d7be242-a072-40ce-b540-95880506cccd",
		Name:         "Dispel",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target instant spell", Instant()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
