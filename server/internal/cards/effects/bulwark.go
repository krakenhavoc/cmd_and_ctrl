package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bulwark — Enchantment {3}{R}{R} (Urza's Saga):
//
//	"At the beginning of your upkeep, this enchantment deals X damage to
//	 target opponent, where X is the number of cards in your hand minus
//	 the number of cards in that player's hand."
//
// The opponent is chosen as the trigger goes on the stack, and both
// hands are counted as it resolves (CR 608.2h). When the opponent holds
// as many cards as you or more, X is zero or less and no damage is dealt
// (CR 107.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "259da825-02f7-4e72-8f44-01683330f457",
		Name:         "Bulwark",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Bulwark — X damage to target opponent",
				sourceDealsComputedDamageToEachLegalTarget(bulwarkX)),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// bulwarkX is the cards in your hand minus the cards in the target
// opponent's hand.
func bulwarkX(ctx *Context) int {
	ts := ctx.LegalTargets()
	if len(ts) == 0 {
		return 0
	}
	return cardsInHand(ctx.Game, ctx.Controller()) - cardsInHand(ctx.Game, ts[0].ID)
}
