package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shoreline Looter — Creature — Rat Rogue {1}{U}, 1/1 (Bloomburrow,
// #4467):
//
//	"This creature can't be blocked.
//	 Threshold — Whenever this creature deals combat damage to a player,
//	 draw a card. Then discard a card unless there are seven or more
//	 cards in your graveyard."
//
// Threshold is an ability word here (CR 207.2c), not a condition on the
// trigger: the Looter always draws. The "unless" is Chakra Meditation's
// shape, a registered "draw, then" body (CR 608.2c) that counts the
// graveyard once the draw is done, as the discard would be made, so a
// draw replaced by a dredge counts the milled cards.
//
// "Can't be blocked" is RestrictSelf(CantBeBlocked), as Blighted Agent's
// is.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c15cad77-e70f-42cc-a9f5-50e1fd8c9c96",
		Name:         "Shoreline Looter",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{RestrictSelf(game.CantBeBlocked)},
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Shoreline Looter — draw a card, then discard a card unless there are seven or more cards in your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					return g.DrawNThenForEffect(item.Controller, 1, game.DrawThen{
						Ref: drawThenDiscardUnlessThreshold, Player: item.Controller, Source: item.SourceCardID, N: 1,
					})
				}),
		},
	})
}

// drawThenDiscardUnlessThreshold is "Then discard a card unless there are
// seven or more cards in your graveyard": the graveyard is counted after
// the draw. The key is an on-disk identity.
var drawThenDiscardUnlessThreshold = game.RegisterDrawThen("discard-unless-threshold", func(g *game.Game, d game.DrawThen) error {
	if fraThresholdMet(g, d.Player) {
		return nil
	}
	g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: d.Player, Source: d.Source, N: d.N})
	return nil
})
