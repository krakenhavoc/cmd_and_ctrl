package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stitcher's Apprentice — {1}{U} Creature — Homunculus 1/2 (Innistrad, S82 OLD
// b446, #5572):
//
//	"{1}{U}, {T}: Create a 2/2 blue Homunculus creature token, then sacrifice a
//	 creature."
//
// The token is made first, so the controller may sacrifice it, the Apprentice
// or any other creature they control; they choose as the ability resolves
// (CR 701.21).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14bf3049-faa3-4c1f-a613-603e977a50a4",
		Name:         "Stitcher's Apprentice",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}, {T}: Create a 2/2 blue Homunculus creature token, then sacrifice a creature.",
			Purpose: game.Purpose{Tokens: 1, Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{1}{U}"), effects.TapCost()),
			Effect: effects.Do(
				effects.CreateToken{Template: effects.TokenCard("2/2 blue Homunculus"), N: 1},
				effects.Instruction(func(ctx *effects.Context) error {
					ctx.Game.PlayerSacrificesForEffect(ctx.Item.SourceCardID, ctx.Controller(), effects.SacrificeSpec("a creature", effects.Creature()),
						"Stitcher's Apprentice — sacrifice a creature")
					return nil
				}),
			),
		}},
	})
}
