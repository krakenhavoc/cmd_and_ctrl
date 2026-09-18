package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gut Shot — Instant {R/P} (EDHREC rank 2260):
//
//	"({R/P} can be paid with either {R} or 2 life.)
//	 Gut Shot deals 1 damage to any target."
//
// The free ping. One damage to any target from the spell as its
// source — Lightning Bolt's shape at one.
//
// The Gitaxian Probe posture, and now with the same ending: #787 made
// the engine pay the Phyrexian symbol with 2 life when the cast
// announces it (CastSpellParams.PhyrexianLife, CR 107.4c), and #916
// gave the cast prompt the button that asks. The free ping is free of
// mana.
func init() {
	Register(Spec{
		OracleID:     "9afb3b6e-4909-4efa-aa79-81c0229411c9",
		Name:         "Gut Shot",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
		},
	})
}
