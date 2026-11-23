package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dive Down — Instant {U} (Foundations, #4465):
//
//	"Target creature you control gets +0/+3 and gains hexproof until end
//	 of turn."
//
// Chase Inspiration's text, built with Adamant Will's body: the boost
// and the keyword land on the still-legal target
// (pumpAndGrantFirstTarget), and nothing happens if it left in
// response.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2f990b54-fbf3-4949-85bb-9ba39710e72a",
		Name:         "Dive Down",
		Purpose:      effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 3, "hexproof")), game.AnswerProtect|game.AnswerPump),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PumpAndGrantFirstTarget(0, 3, "Dive Down — +0/+3 and hexproof", "hexproof")(ctx.Game, item)
		},
	})
}
