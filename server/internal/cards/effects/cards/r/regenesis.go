package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Regenesis — Instant {3}{G}{G} (Ravnica Allegiance):
//
//	"Return up to two target permanent cards from your graveyard to your hand."
//
// March of the Returned's body over permanent cards: each target is re-checked
// as the spell resolves (CR 608.2b) and the ones still in the graveyard return.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14bc2a2e-a445-4785-9c88-b9ad169c39bc",
		Name:         "Regenesis",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("up to two target permanent cards from your graveyard", effects.YouOwn(), effects.Permanent()).WithCount(0, 2),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ReturnLegalGraveyardTargetsToHand)),
	})
}
