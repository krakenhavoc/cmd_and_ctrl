package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unmake the Graves — Instant {4}{B} (Magic 2015, #5556):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Return up to two target creature cards from your graveyard to your
//	 hand."
//
// Fight On!'s return (returnLegalGraveyardTargetsToHand) with the convoke
// TapCost: up to two, so it may name none, and a card that left in
// response stays where it went.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d9b99c58-8e2e-496c-8288-eda934f49162",
		Name:         "Unmake the Graves",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("up to two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(0, 2),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand)),
	})
}
