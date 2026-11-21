package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mysterio's Mirage — {2}{U} Enchantment (Marvel Super Heroes Commander,
// #4455):
//
//	"At the beginning of your end step, if you discarded a card this turn,
//	 create a 3/3 blue Illusion Villain creature token."
//
// The "if" is intervening (CR 603.4): it is checked as the step begins and
// again as the ability resolves, against the player's turn tally of cards
// discarded (one per card, CR 701.9a). It triggers once however many were
// discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60733a3a-dfc9-429a-af2d-f0ce9a58d533",
		Name:         "Mysterio's Mirage",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "you discarded a card this turn",
				Test: func(ctx *effects.Context) bool { return mscM9DiscardedThisTurn(ctx.Game, ctx.Controller()) },
			}, effects.On(game.EventBeginEndStep,
				effects.AllOf(effects.ByYou, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return mscM9DiscardedThisTurn(g, source.Controller)
				}),
				"Mysterio's Mirage — create a 3/3 blue Illusion Villain creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("3/3 blue Illusion Villain"), N: 1}))),
		},
	})
}

// mscM9DiscardedThisTurn is "you discarded a card this turn".
func mscM9DiscardedThisTurn(g *game.Game, player uuid.UUID) bool {
	return g.TurnTallyFor(player).CardsDiscarded > 0
}
