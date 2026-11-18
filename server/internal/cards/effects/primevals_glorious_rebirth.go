package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primevals' Glorious Rebirth — Legendary Sorcery {5}{W}{B}:
//
//	"(You may cast a legendary sorcery only if you control a legendary
//	 creature or planeswalker.)
//	 Return all legendary permanent cards from your graveyard to the
//	 battlefield."
//
// The cast gate is checked once, at announce (CR 205.4e; LegendarySorcery).
// The cards are read before anything moves and enter together under their
// owner's control, the caster's ("your graveyard"), so each sees the others
// enter (CR 603.6a). A legendary instant or sorcery card is not a permanent
// card and stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:           "8de8027b-072a-474b-b76b-dc49c417cc55",
		Name:               "Primevals' Glorious Rebirth",
		Completeness:       CompletenessFull,
		CastCondition:      LegendarySorcery(),
		CastConditionLabel: LegendarySorceryLabel,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return c.IsPermanent() && c.IsLegendary()
				}),
			}.Apply(ctx)
		},
	})
}
