package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sacred Excavation — {3}{U} Sorcery (Amonkhet, S82 old batch 248):
//
//	"Return up to two target cards with cycling from your graveyard to your
//	 hand."
//
// "A card with cycling" is a card with a cycling ability or typecycling
// (CR 702.29), read off the card in the graveyard. Each target still legal as
// the spell resolves returns (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5d8ff6a0-7037-4231-b6a9-ad97260bb267",
		Name:         "Sacred Excavation",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets: effects.TargetCardInGraveyard("up to two target cards with cycling from your graveyard",
			effects.YouOwn(), effects.GapB158HasCyclingAbility).WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
