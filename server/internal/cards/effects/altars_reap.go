package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Altar's Reap — Instant {1}{B}:
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Draw two cards."
//
// Village Rites at one mana more — printed years earlier, and still
// played in lists that want a second copy of the effect. Identical
// machinery; kept as its own card because the extra mana genuinely
// matters to the decks running both.
func init() {
	Register(Spec{
		OracleID:       "6a125750-2b8c-4f9d-8173-ac8d14c91ddb",
		Name:           "Altar's Reap",
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
		},
	})
}
