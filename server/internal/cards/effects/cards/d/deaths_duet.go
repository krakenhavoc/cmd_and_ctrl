package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death's Duet — Sorcery {2}{B} (Exodus, #5626):
//
//	"Return two target creature cards from your graveyard to your hand."
//
// Blood Beckoning's kicked clause as the only one: exactly two targets
// (WithCount(2, 2)), so it can't be cast with fewer than two creature
// cards in your graveyard (CR 601.2c). Each target is re-checked at
// resolution (CR 608.2b), and the shared returnLegalGraveyardTargetsToHand
// returns the ones still legal, so one that left in response is skipped
// and the other still comes back.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "28d24add-da3a-4697-8fd2-5106c9225c0e",
		Name:         "Death's Duet",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		Targets:      effects.TargetCardInGraveyard("two target creature cards from your graveyard", effects.YouOwn(), effects.Creature()).WithCount(2, 2),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnLegalGraveyardTargetsToHand(ctx)
		},
	})
}
