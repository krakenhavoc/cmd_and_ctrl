package cardss

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sanctum of Calm Waters — Legendary Enchantment — Shrine {3}{U} (Core Set
// 2021, #5511):
//
//	"At the beginning of your first main phase, you may draw X cards, where X
//	 is the number of Shrines you control. If you do, discard a card."
//
// "Your first main phase" is the precombat main. X is counted as the player
// answers, then the draws are made and the discard follows them. A player a
// draw limit stops from drawing that many is not asked and does neither
// (CR 121.2b).
//
// No simplification.
func init() {
	const label = "Sanctum of Calm Waters — you may draw X cards, where X is the number of Shrines you control; if you do, discard a card"
	effects.Register(effects.Spec{
		OracleID:     "1cd1e3a8-e09d-4b61-8244-92b317132d6a",
		Name:         "Sanctum of Calm Waters",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			effects.AtYourPrecombatMain(label, func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				x := effects.S82OldB105ShrinesYouControl(g, ctx.Controller())
				return effects.MayChoice{
					Question: "Sanctum of Calm Waters — draw " + effects.NumberWord(x) + " cards, then discard a card?",
					Draws:    x,
					Then:     sanctumOfCalmWatersAnswer,
					Word:     strconv.Itoa(x),
				}.Apply(ctx)
			}),
		},
	})
}

// sanctumOfCalmWatersAnswer draws X cards (the Word) and then discards one.
var sanctumOfCalmWatersAnswer = effects.MayYes("confirm/s82-old-b105-sanctum-of-calm-waters", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if err := (effects.DrawCards{Player: r.Chooser, N: effects.WordInt(r)}).Apply(ctx); err != nil {
		return err
	}
	ctx.Game.QueueDiscardChoiceForEffect(game.DiscardPrompt{
		Player:   r.Chooser,
		Source:   ctx.Source(),
		N:        1,
		Question: "Sanctum of Calm Waters — discard a card",
	})
	return nil
})
