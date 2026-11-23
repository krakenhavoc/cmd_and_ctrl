package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sam's Desperate Rescue — Sorcery {B}:
//
//	"Return target creature card from your graveyard to your hand. The
//	 Ring tempts you."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5a3336f1-614a-4b3f-8752-522a34d0417a",
		Name:         "Sam's Desperate Rescue",
		Purpose:      game.Purpose{RingTempts: 1},
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := effects.ReturnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			return effects.TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
