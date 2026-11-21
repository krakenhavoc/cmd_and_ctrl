package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fetid Gargantua — Creature — Horror {4}{B}, 4/4 (Modern Horizons 3):
//
//	"{2}{B}: Adapt 2.
//	 Whenever one or more +1/+1 counters are put on this creature, you may
//	 draw two cards. If you do, you lose 2 life."
func init() {
	const label = "Fetid Gargantua — you may draw two cards. If you do, you lose 2 life"
	effects.Register(effects.Spec{
		OracleID:     "b8c35a80-2c72-49ff-af5c-e8885c97f6fc",
		Name:         "Fetid Gargantua",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Adapt 2.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.ManaCost("{2}{B}"),
			Effect:  func(g *game.Game, item *game.StackItem) error { return effects.AdaptN(g, item, 2) },
		}},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventCounterPlaced, effects.GapB49CountersPlacedOnThis, label,
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayChoice{
						Question: label + "?",
						Draws:    2,
						Then:     fetidGargantuaAnswer,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var fetidGargantuaAnswer = effects.MayYes("confirm/fetid-gargantua-draw", func(ctx *effects.Context, r game.ConfirmAnswer) error {
	if err := (effects.DrawCards{Player: r.Chooser, N: 2}).Apply(ctx); err != nil {
		return err
	}
	return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), r.Chooser, -2)
})
