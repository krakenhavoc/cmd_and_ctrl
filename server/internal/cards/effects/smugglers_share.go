package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Smuggler's Share — Enchantment {2}{W} (EDHREC rank 1687):
//
//	"At the beginning of each end step, draw a card for each opponent
//	 who drew two or more cards this turn, then create a Treasure
//	 token for each opponent who had two or more lands enter the
//	 battlefield under their control this turn."
//
// White's catch-up enchantment: it taxes the table for drawing and
// ramping harder than you. EACH end step, so it asks on every turn;
// not an intervening-if, so it triggers every end step and does
// nothing when nobody qualified. Cards first, then Treasures, in
// printed order.
//
// Both counts are cells of the engine's per-turn tally
// (b15CardsDrawnThisTurn, b15LandsEnteredThisTurn). Draws are one
// event per card with the drawer stamped; lands are looked up where
// they sit now for their type and controller, so a fetchland that
// entered and sacrificed itself still counts, as printed, and a land
// whose control has since changed counts for its new controller —
// weaker for its old one, never stronger.
//
// No simplification.
// smugglersShareCounts are the cards to draw (opponents who drew two or more
// this turn) and the Treasures to make (opponents who played two or more
// lands). Both read the turn's record, so each step reads its own.
func smugglersShareCounts(ctx *Context) (cards, treasures int) {
	drawn, lands := b15CardsDrawnThisTurn(ctx.Game), b15LandsEnteredThisTurn(ctx.Game)
	for _, opp := range ctx.Opponents() {
		if drawn[opp] >= 2 {
			cards++
		}
		if lands[opp] >= 2 {
			treasures++
		}
	}
	return cards, treasures
}

func init() {
	Register(Spec{
		OracleID:     "17b29350-4f37-4552-8192-4856b15345f9",
		Name:         "Smuggler's Share",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15EndStepBegan(ev)
			}, "Smuggler's Share — a card per greedy opponent, a Treasure per ramping one", Do(
				// The Treasures wait for a prompt the draws ask (CR 608.2c, #4155). Both
				// counts are about the opponents, so the draw does not change the second.
				Instruction(func(ctx *Context) error {
					cards, _ := smugglersShareCounts(ctx)
					return DrawCards{N: cards}.Apply(ctx)
				}),
				Instruction(func(ctx *Context) error {
					_, treasures := smugglersShareCounts(ctx)
					if treasures == 0 {
						return nil
					}
					return CreateToken{Template: TreasureToken(), N: treasures}.Apply(ctx)
				}),
			)),
		},
	})
}
