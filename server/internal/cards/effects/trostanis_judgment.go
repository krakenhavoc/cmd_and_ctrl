package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trostani's Judgment — Instant {5}{W}:
//
//	"Exile target creature, then populate. (Create a token that's a
//	 copy of a creature token you control.)"
//
// The populate is the exile's continuation, not the next line: the
// exile opens the CR 614 replacement window and may stop to ask, and
// the populate must see the board after the creature has left (so
// exiling your own token does not leave it as a candidate). If the
// target is gone by resolution the spell does nothing at all
// (CR 608.2b), populate included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "97cf544e-ffbf-4730-8cd2-4f1d5be933f2",
		Name:         "Trostani's Judgment",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return ExileTarget{
				Target: item.Targets[0].ID,
				Then: func(c *Context, _ bool) error {
					return Populate{}.Apply(c)
				},
			}.Apply(ctx)
		},
	})
}
