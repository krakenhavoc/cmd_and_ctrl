package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sorcerer's Strongbox — Artifact {4} (Archenemy, #5580):
//
//	"{2}, {T}: Flip a coin. If you win the flip, sacrifice this artifact
//	 and draw three cards."
//
// The Gold Saucer's keyed coin call (CR 705.2), with the ability carried
// on the prompt so the payoff reads the Strongbox it came from. A lost flip
// does nothing. A won flip sacrifices the Strongbox only while it is still
// that object (CR 400.7) and still under the control of the ability's
// controller, since a player can't sacrifice a permanent they don't
// control (CR 701.21a, the ruling); the three cards are drawn either way.
// The draw is the sacrifice's keyed tail, so it happens after the
// Strongbox has gone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c4f58768-35b2-4e17-9129-c9e39b8ed339",
		Name:         "Sorcerer's Strongbox",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Flip a coin. If you win the flip, sacrifice this artifact and draw three cards.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				g.FlipCoinForEffect(game.CoinFlipSpec{
					Flipper:  item.Controller,
					Source:   item.SourceCardID,
					Question: "Sorcerer's Strongbox — call the coin flip",
					ThenKey:  sorcerersStrongboxFlipped,
					Asked:    game.NewAskedItem(item),
				})
				return nil
			},
		}},
	})
}

// sorcerersStrongboxFlipped is the settled call: a won flip sacrifices
// the Strongbox if you still control that object, then draws three. An
// on-disk key: never renamed, never reused.
var sorcerersStrongboxFlipped = effects.CoinFlipThen("coin-flip/sorcerers-strongbox", func(ctx *effects.Context, r game.CoinFlipped) error {
	if !effects.WonFirstFlip(r) {
		return nil
	}
	if !sorcerersStrongboxStillYours(ctx) {
		return sorcerersStrongboxDrawThree(ctx)
	}
	return effects.SacrificePermanent{Target: ctx.Item.SourceCardID, ThenTail: ctx.ItemTail(sorcerersStrongboxDraw)}.Apply(ctx)
})

// sorcerersStrongboxStillYours reports whether the ability's source is
// still the same permanent and its controller still controls it.
func sorcerersStrongboxStillYours(ctx *effects.Context) bool {
	if !effects.SourceIsStillThisPermanent(ctx.Game, ctx.Item) {
		return false
	}
	c, ok := ctx.Game.LookupCardForEffect(ctx.Item.SourceCardID)
	return ok && c.Controller == ctx.Controller()
}

// sorcerersStrongboxDraw is "and draw three cards" after the sacrifice,
// whether or not it landed. An on-disk key (#4253 slice 10a).
var sorcerersStrongboxDraw = effects.CardTail("then/sorcerers-strongbox-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return sorcerersStrongboxDrawThree(ctx)
})

func sorcerersStrongboxDrawThree(ctx *effects.Context) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
}
