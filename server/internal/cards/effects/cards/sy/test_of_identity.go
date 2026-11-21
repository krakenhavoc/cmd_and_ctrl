package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Test of Identity — Instant {1}{R} (Star Trek, #4973):
//
//	"Choose one —
//	 • I Am Klingon — Target Klingon gets +1/+1 and gains double strike
//	   until end of turn.
//	 • I Am Starfleet — Discard a card. If you do, draw two cards."
//
// "I Am Klingon" and "I Am Starfleet" are flavor words (CR 207.2d): they
// name the modes and have no rules meaning. Mode one targets a permanent
// with the Klingon subtype, read through the layers, so a changeling
// counts (CR 702.73a); the pump and the double strike end at cleanup and
// nothing happens if the target is gone or illegal as the spell resolves
// (CR 608.2b). Mode two is the mandatory discard with the shared
// "if you do, draw two" continuation: the card is chosen as the spell
// resolves, and with an empty hand nothing is discarded and nothing is
// drawn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8cb700bb-daa5-4a26-bb84-e22bc11bdaff",
		Name:         "Test of Identity",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("I Am Klingon — Target Klingon gets +1/+1 and gains double strike until end of turn.",
				effects.TargetPermanent("target Klingon", effects.Subtype("Klingon"))),
				game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 1, 1, "double strike")), Answers: game.AnswerCombatGrant}),
			effects.ModeWithPurpose(effects.Mode("I Am Starfleet — Discard a card. If you do, draw two cards."),
				game.Purpose{Answers: game.AnswerValue}),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			switch {
			case ctx.HasMode(0):
				return effects.PumpAndGrantFirstTarget(1, 1, "Test of Identity — +1/+1 and double strike", "double strike")(ctx.Game, item)
			case ctx.HasMode(1):
				return ctx.Game.PlayerDiscardsThenKeyedForEffect(game.DiscardPrompt{
					Player: ctx.Controller(),
					Source: ctx.Source(),
					N:      1,
				}, effects.DrawTwoIfDiscarded.Then, nil)
			}
			return nil
		},
	})
}
