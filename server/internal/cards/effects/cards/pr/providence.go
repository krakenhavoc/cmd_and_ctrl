package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Providence — Sorcery {5}{W}{W} (Eldritch Moon, S82 OLD, #5541):
//
//	"You may reveal this card from your opening hand. If you do, at the
//	 beginning of the first upkeep, your life total becomes 26.
//	 Your life total becomes 26."
//
// The opening-hand reveal is Chancellor of the Dross' seam
// (RevealFromOpeningHand at game.RevealAtTheFirstUpkeep, CR 103.6): the card
// stays in the hand, revealed, and a delayed trigger controlled by its owner
// fires at the first upkeep of the game, whoever's it is. The life change is
// Oketra's Last Mercy's (CR 119.5): life gained or lost for the difference,
// so lifegain and life-loss triggers see it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43b164d5-2bff-4a03-a4fa-16ffad7e3164",
		Name:         "Providence",
		Completeness: effects.CompletenessFull,
		OpeningHand: effects.RevealFromOpeningHand(game.RevealAtTheFirstUpkeep,
			"Providence — your life total becomes 26", providenceFirstUpkeep),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.B31LifeBecomes(ctx, ctx.Controller(), 26)
		},
	})
}

var providenceFirstUpkeep = game.SimpleDelayedBody("providence/first-upkeep-life", func(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	return effects.B31LifeBecomes(ctx, ctx.Controller(), 26)
})
