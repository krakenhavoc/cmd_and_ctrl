package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sam's Desperate Rescue — Sorcery {B}:
//
//	"Return target creature card from your graveyard to your hand. The
//	 Ring tempts you."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5a3336f1-614a-4b3f-8752-522a34d0417a",
		Name:         "Sam's Desperate Rescue",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target creature card from your graveyard", YouOwn(), Creature()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if err := returnLegalGraveyardTargetsToHand(ctx); err != nil {
				return err
			}
			return TheRingTemptsYou{}.Apply(ctx)
		},
	})
}
