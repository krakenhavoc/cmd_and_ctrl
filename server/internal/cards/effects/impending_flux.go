package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Impending Flux — Sorcery {2}{R} (Doctor Who):
//
//	"Paradox — Impending Flux deals X damage to each opponent and each
//	 creature they control, where X is 1 plus the number of spells you've cast
//	 from anywhere other than your hand this turn.
//	 Foretell {1}{R}{R} (During your turn, you may pay {2} and exile this card
//	 from your hand face down. Cast it on a later turn for its foretell cost.)"
//
// X is read as the spell resolves, off the cast tally's NotFromHand (#5235):
// the spells its caster cast this turn from anywhere but the hand. Impending
// Flux itself counts when it was cast from exile by its foretell cost, since it
// is a spell cast this turn from outside the hand. The damage is one damage
// instance to each opponent and each creature they control. Foretell is the
// engine's special action.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "7089512e-f9b6-4d8e-bccc-29e6785cc041",
		Name:           "Impending Flux",
		Completeness:   CompletenessFull,
		Purpose:        game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, OpponentsOnly: true, Partial: true}},
		SpecialActions: []game.SpecialAction{Foretell("{1}{R}{R}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			x := 1 + ctx.Game.CastTallyFor(item.Controller).NotFromHand
			return damageEachMatchingAndEachOpponent(ctx, And(Creature(), OpponentControls()), x)
		},
	})
}
