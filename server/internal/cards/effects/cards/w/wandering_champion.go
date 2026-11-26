package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Wandering Champion — {1}{W} Creature — Human Monk 3/1 (Fate Reforged):
//
//	"Whenever this creature deals combat damage to a player, if you control a
//	 blue or red permanent, you may discard a card. If you do, draw a card."
//
// An intervening "if" (CR 603.4): a blue or red permanent is checked as the
// damage is dealt and again as the trigger resolves, on effective colors. The
// draw follows only a card actually discarded.
func init() {
	const label = "Wandering Champion — you may discard a card. If you do, draw a card"
	blueOrRed := effects.CondBoard{
		Name: "you control a blue or red permanent",
		Test: func(ctx *effects.Context) bool {
			blue, red := ctx.TextColor("U"), ctx.TextColor("R")
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == ctx.Controller() && (c.HasColor(blue) || c.HasColor(red)) {
					return true
				}
			}
			return false
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "54af9e23-459e-470d-9b0e-e91a0fa45c98",
		Name:         "Wandering Champion",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(blueOrRed, effects.WheneverThisDealsCombatDamageToAPlayer(label, func(g *game.Game, item *game.StackItem) error {
				return effects.B39MayDiscardThenDraw(1, false, label, effects.DrawOneIfDiscarded)(effects.NewContext(g, item))
			})),
		},
	})
}
