package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gretchen Titchwillow — Legendary Creature — Halfling Druid {G}{U}, 0/4
// (Adventures in the Forgotten Realms):
//
//	"{2}{G}{U}: Draw a card. You may put a land card from your hand onto
//	 the battlefield."
//
// Two instructions in order: the draw first, so the card just drawn can
// be the land (CR 608.2c); then Budoka Gardener's optional land drop.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cedeea30-b580-4ea7-bbc3-7358fed8ae61",
		Name:         "Gretchen Titchwillow",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}{U}: Draw a card. You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}{U}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				if err := (effects.DrawCards{Player: item.Controller, N: 1}).Apply(ctx); err != nil {
					return err
				}
				return effects.MayPutALandFromHand("Gretchen Titchwillow").Apply(ctx)
			},
		}},
	})
}
