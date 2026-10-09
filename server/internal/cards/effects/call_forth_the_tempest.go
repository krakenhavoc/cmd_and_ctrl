package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Call Forth the Tempest — Sorcery {5}{R}{R}{R}:
//
//	"Cascade, cascade (When you cast this spell, exile cards from the
//	 top of your library until you exile a nonland card that costs
//	 less. You may cast it without paying its mana cost. Put the
//	 exiled cards on the bottom of your library in a random order.
//	 Then do it again.)
//	 Call Forth the Tempest deals damage to each creature your
//	 opponents control equal to the total mana value of other spells
//	 you've cast this turn."
//
// Two Cascade() triggers, each its own instance (CR 702.85c). The
// spells they cast resolve first and are already in the cast tally
// when this resolves, which is the point of the card.
//
// The amount is the tally's running mana value (#2743) minus this
// spell's own. A copy was never cast, so it subtracts nothing, and
// "other spells" then counts the original too.
//
// No Purpose sweep: a damage sweep declares a fixed amount or X (ADR
// 0126 §6), and this amount is neither until the spell resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0623c51a-e829-413f-a3a0-817c0902821e",
		Name:         "Call Forth the Tempest",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Cascade(), Cascade()},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), OpponentControls()),
				manaValueOfOtherSpellsYouCastThisTurn(ctx, item))
		},
	})
}

// manaValueOfOtherSpellsYouCastThisTurn is "the total mana value of
// other spells you've cast this turn", read by the resolving spell
// `item`: its controller's tally, less its own contribution if it was
// cast rather than copied (#2743).
func manaValueOfOtherSpellsYouCastThisTurn(ctx *Context, item *game.StackItem) int {
	total := ctx.Game.CastTallyFor(ctx.Controller()).ManaValue
	if item.IsCopy {
		return total
	}
	for i := range ctx.Game.Stack.Cards {
		if sc := &ctx.Game.Stack.Cards[i]; sc.InstanceID == item.SourceCardID {
			if mv, ok := ctx.Game.CastManaValueForEffect(sc); ok {
				total -= mv
			}
			break
		}
	}
	return max(total, 0)
}
