package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Axis of Mortality — Enchantment {4}{W}{W} :
//
//	"At the beginning of your upkeep, you may have two target
//	 players exchange life totals."
//
// The group-hug politics card: every upkeep you hand two seats each
// other's life total, and nothing says one of them has to be you.
// The clause targets TWO PLAYERS, not "you and target opponent", so
// the controller can make two opponents trade and keep their own
// total out of it — which is most of the card's table presence.
//
// Three things the shape gets right on its own:
//
//   - "You may" is CR 603.4, so the trigger asks before it builds.
//     Declining is a real line: on a turn where the swap would help
//     the wrong seat, you simply do not swap.
//   - Both target slots must be filled at announce, and the engine
//     re-checks each at resolution (CR 608.2b). With a seat gone in
//     response there is only one legal target left and nothing is
//     exchanged — an exchange is all or nothing, not "the survivor
//     keeps what they have and gains nothing".
//   - The exchange itself is exchangeLifeTotals: both totals read
//     before either is written, each half through the CR 614 window,
//     so a lifegain or life-loss payoff at either seat sees it.
//
// A three-player table with two identical life totals among the
// chosen pair exchanges nothing, correctly and silently.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "158cb888-a115-4d50-a82f-746622f0578e",
		Name:         "Axis of Mortality",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					AtYourUpkeep("Axis of Mortality — two target players exchange life totals",
						axisOfMortalityExchange),
					"Axis of Mortality — have two target players exchange life totals?"),
				TargetPlayer("two target players").WithCount(2, 2)),
		},
	})
}

// axisOfMortalityExchange swaps the two chosen seats' life totals.
// Both slots must still be legal: CR 701's exchange is all or
// nothing, so one surviving target is nothing to exchange with.
func axisOfMortalityExchange(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	legal := ctx.LegalTargets()
	if len(legal) != 2 || legal[0].Kind != game.TargetPlayer || legal[1].Kind != game.TargetPlayer {
		return nil
	}
	return exchangeLifeTotals(ctx, legal[0].ID, legal[1].ID)
}
