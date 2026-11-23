package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inevitable Defeat — Instant {1}{R}{W}{B} (Tarkir: Dragonstorm, #4463):
//
//	"This spell can't be countered.
//	 Exile target nonland permanent. Its controller loses 3 life and you
//	 gain 3 life."
//
// Anguished Unmaking's exile with Bitter Downfall's victim: the
// permanent's controller is read before the exile, because once it has
// left the battlefield there is nothing left to ask whose it was. The life
// swing does not depend on the exile taking, as printed. With the target
// gone the whole spell does nothing (CR 608.2b). The uncounterable half is
// Spec.CantBeCountered.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d68439b5-7011-4b1d-a7da-14ef0f4d9994",
		Name:            "Inevitable Defeat",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Purpose:         game.Purpose{Answers: game.AnswerRemove},
		Targets:         effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok {
					continue
				}
				victim := c.Controller
				if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
				if err := ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), victim, -3); err != nil {
					return err
				}
				return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
			}
			return nil
		},
	})
}
