package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Surge of Brilliance — Instant {1}{U} (Doctor Who):
//
//	"Paradox — Draw a card for each spell you've cast this turn from
//	 anywhere other than your hand.
//	 Foretell {1}{U} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell
//	 cost.)"
//
// The count is the cast tally's NotFromHand (#5235): the spells the caster
// cast this turn from a graveyard, exile, the command zone or a library,
// counted as each was cast. Surge itself is included when it is cast from
// exile by its foretell cost, since it is a spell cast this turn from
// anywhere other than the hand. Foretell is the engine's special action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "4d720849-56cd-4b78-b9a7-7b9484702f0f",
		Name:           "Surge of Brilliance",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Draws: 1},
		SpecialActions: []game.SpecialAction{Foretell("{1}{U}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			n := ctx.Game.CastTallyFor(item.Controller).NotFromHand
			if n <= 0 {
				return nil
			}
			return DrawCards{N: n}.Apply(ctx)
		},
	})
}
