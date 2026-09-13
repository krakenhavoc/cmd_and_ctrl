package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thought Scour — Instant {U} (EDHREC rank 1785):
//
//	"Target player mills two cards.
//	 Draw a card."
//
// The one-mana self-mill cantrip. Glimpse the Unthinkable's target
// clause, a mill of two, then the caster draws — in the printed
// order, so a Thought Scour aimed at yourself mills before it draws
// and the card drawn is the third from the top.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "83101ba8-a569-4827-8c53-9ca0dfcd59a7",
		Name:         "Thought Scour",
		Completeness: CompletenessFull,
		Targets:      TargetPlayer("target player"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetPlayer {
					continue
				}
				if err := (MillCards{Player: t.ID, N: 2}).Apply(ctx); err != nil {
					return err
				}
			}
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
