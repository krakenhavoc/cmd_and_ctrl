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
func init() {
	Register(Spec{
		OracleID:     "17b29350-4f37-4552-8192-4856b15345f9",
		Name:         "Smuggler's Share",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return b15EndStepBegan(ev)
			}, "Smuggler's Share — a card per greedy opponent, a Treasure per ramping one", func(g *game.Game, item *game.StackItem) error {
				ctx := NewContext(g, item)
				drawn, lands := b15CardsDrawnThisTurn(g), b15LandsEnteredThisTurn(g)
				cards, treasures := 0, 0
				for _, opp := range ctx.Opponents() {
					if drawn[opp] >= 2 {
						cards++
					}
					if lands[opp] >= 2 {
						treasures++
					}
				}
				if err := (DrawCards{Player: item.Controller, N: cards}).Apply(ctx); err != nil {
					return err
				}
				if treasures == 0 {
					return nil
				}
				return CreateToken{Controller: item.Controller, Template: TreasureToken(), N: treasures}.Apply(ctx)
			}),
		},
	})
}
