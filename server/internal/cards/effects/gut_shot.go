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
// The Gitaxian Probe posture: since #787 the engine pays the
// Phyrexian symbol with 2 life when the cast announces it
// (CastSpellParams.PhyrexianLife, CR 107.4c), and the board has no
// button that asks. Clicking the card from hand still charges {R}. A
// client seam, not a card file's.
func init() {
	Register(Spec{
		OracleID:     "9afb3b6e-4909-4efa-aa79-81c0229411c9",
		Name:         "Gut Shot",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The board has no button for the Phyrexian symbol yet — casting it from hand pays {R}, not 2 life. The engine accepts the life payment."},
		Targets:      TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: item.Targets[0].ID, Amount: 1}.Apply(ctx)
		},
	})
}
