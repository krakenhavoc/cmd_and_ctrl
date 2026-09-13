package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Counterspell — "Counter target spell." The canonical counter-
// magic card. Exit-criteria card: an opponent's Lightning Bolt on
// the stack, cast Counterspell targeting it, pass priority, Bolt
// goes to opponent's graveyard without damage.
//
// Sandbox predicate note: Counterspell targets any spell (no
// "non-creature" gate — that's S20 smart-cast territory). The
// announce-time UI lets the caster pick any stack item; the
// CounterTarget primitive doesn't enforce card-type restrictions.
func init() {
	Register(Spec{
		OracleID:     "cc187110-1148-4090-bbb8-e205694a39f5",
		Name:         "Counterspell",
		Completeness: CompletenessFull,
		Targets:      TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return CounterTarget{StackID: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
