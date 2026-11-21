package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Platoon Dispenser — Artifact Creature — Construct {5}, 4/6 (The Brothers'
// War, #5482):
//
//	"At the beginning of your end step, if you control two or more other
//	 creatures, draw a card.
//	 {3}{W}: Create a 1/1 colorless Soldier artifact creature token.
//	 Unearth {2}{W}{W} (…)"
//
// The end-step draw is an intervening "if" (CR 603.4): the count of OTHER
// creatures you control is read as the step begins and again as the ability
// resolves. The token is Third Path Iconoclast's. Unearth is the shared
// keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "339c39d9-7e10-458e-b602-4f0550840436",
		Name:         "Platoon Dispenser",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{Name: "you control two or more other creatures", Test: func(ctx *effects.Context) bool {
				n := 0
				for _, id := range effects.PermanentsControlledByMatching(ctx.Game, ctx.Controller(), effects.Creature()) {
					if id != ctx.Source() {
						n++
					}
				}
				return n >= 2
			}}, effects.AtYourEndStep("Platoon Dispenser — draw a card", effects.Do(effects.DrawCards{N: 1}))),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{3}{W}: Create a 1/1 colorless Soldier artifact creature token.",
				Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerMakesBlocker},
				Cost:    effects.ManaCost("{3}{W}"),
				Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 colorless Soldier artifact"), N: 1}),
			},
			effects.Unearth("{2}{W}{W}"),
		},
	})
}
