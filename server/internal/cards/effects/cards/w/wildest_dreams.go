package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildest Dreams — Sorcery {X}{X}{G} (Kaladesh, #5539):
//
//	"Return X target cards from your graveyard to your hand. Exile Wildest
//	 Dreams."
//
// Back in Town's X targets (TargetsCountedByX): exactly X, chosen as the spell
// is cast, and every one still legal as it resolves comes back (CR 608.2b).
// Each {X} counts toward the cost, so X is paid twice. Then the spell exiles
// itself (Temporal Mastery's move), so it never reaches the graveyard it just
// emptied. With X at 0 it has no targets, returns nothing, and is still
// exiled; if every target is illegal it does not resolve at all and goes to
// the graveyard like any countered-on-resolution spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a25a72e-5603-4b50-ac20-36b286cc1a2b",
		Name:         "Wildest Dreams",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetsCountedByX(effects.TargetCardInGraveyard("X target cards from your graveyard", effects.YouOwn())),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand),
			effects.Instruction(func(ctx *effects.Context) error {
				return ctx.Game.ExileCardForEffect(ctx.Source())
			}),
		),
	})
}
