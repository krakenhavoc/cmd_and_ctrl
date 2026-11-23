package cardsth

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tivash, Gloom Summoner — {4}{B} Legendary Creature — Human Warlock 4/4 (Jumpstart 2022, #5948):
//
//	"Lifelink
//	 At the beginning of your end step, if you gained life this turn, you
//	 may pay X life, where X is the amount of life you gained this turn.
//	 If you do, create an X/X black Demon creature token with flying."
//
// The intervening "if" (CR 603.4) is checked as the end step begins and
// again as the trigger resolves; X is the life gained this turn when it
// resolves. Paying life is a cost, so it is re-checked as the answer
// arrives (CR 119.4) and the token is made only if it was paid.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b5641fcf-135e-43e7-a547-344fac61e279",
		Name:            "Tivash, Gloom Summoner",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, effects.AllOf(effects.ByYou, tivashGainedLife),
				"Tivash, Gloom Summoner — you may pay X life; if you do, create an X/X black Demon creature token with flying",
				func(g *game.Game, item *game.StackItem) error {
					x := effects.B15LifeGainedThisTurn(g, item.Controller)
					if x <= 0 {
						return nil
					}
					return effects.MayChoice{
						Question: "Tivash, Gloom Summoner — pay " + strconv.Itoa(x) + " life to create a " + strconv.Itoa(x) + "/" + strconv.Itoa(x) + " Demon?",
						YesLabel: "Pay " + strconv.Itoa(x) + " life",
						NoLabel:  "Decline",
						LifeCost: x,
						Then:     tivashPayLifeForDemon,
						Word:     strconv.Itoa(x),
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

func tivashGainedLife(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	return effects.B15LifeGainedThisTurn(g, source.Controller) > 0
}

var tivashPayLifeForDemon = effects.MayYes("confirm/tivash-gloom-summoner-demon", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	x := effects.WordInt(r)
	who := ctx.Game.PlayerByIDForEffect(r.Chooser)
	if x <= 0 || who == nil || who.Eliminated || who.Life < x {
		return nil
	}
	if err := ctx.Game.PayLifeForEffect(ctx.Source(), r.Chooser, x); err != nil {
		return err
	}
	return effects.CreateToken{Controller: r.Chooser, N: 1, Template: game.Card{
		Name: "Demon", TypeLine: "Token Creature — Demon", Power: x, Toughness: x,
		Colors: []string{"B"}, Keywords: []string{"flying"},
	}}.Apply(ctx)
})
