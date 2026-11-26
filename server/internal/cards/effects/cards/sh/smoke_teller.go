package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Smoke Teller — Creature — Human Shaman {1}{G}, 2/2 (Khans of Tarkir):
//
//	"{1}{U}: Look at target face-down creature."
//
// The target is any face-down creature on the battlefield, yours or not, and
// is re-checked as the ability resolves (CR 608.2b): one turned face up in
// response is no longer a legal target and nothing is seen. The look is
// private (game.LookAtFaceDownCreatureForEffect, #5554): the activator
// becomes a knower of that one creature and nobody else is told.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "387a5937-cdbd-4e8e-8efb-4d444aa9cbc9",
		Name:         "Smoke Teller",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Look at target face-down creature.",
			Cost:    effects.ManaCost("{1}{U}"),
			Targets: effects.TargetCreature("target face-down creature", effects.FaceDown()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, t := range ctx.LegalTargets() {
					if t.Kind == game.TargetCard {
						g.LookAtFaceDownCreatureForEffect(ctx.Controller(), t.ID)
					}
				}
				return nil
			},
		}},
	})
}
