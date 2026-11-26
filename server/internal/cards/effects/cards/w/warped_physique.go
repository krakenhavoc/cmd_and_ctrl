package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Warped Physique — Instant {U}{B} (Dragon's Maze, S82 OLD batch 403, #5564):
//
//	"Target creature gets +X/-X until end of turn, where X is the number of cards
//	 in your hand."
//
// Flowstone Slide's +X/-X on one target: X is the number of cards in your hand as
// the spell resolves, after the spell itself has left it (CR 608.2n), fixed for
// the effect (CR 611.2c). With an empty hand X is zero and nothing changes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f45fdef0-15b1-472b-9485-465c49a48a85",
		Name:         "Warped Physique",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			you := ctx.Game.PlayerByIDForEffect(item.Controller)
			if you == nil {
				return nil
			}
			x := you.Hand.Size()
			return effects.BoostUntilEOT{Target: id, Power: x, Toughness: -x, Label: "Warped Physique — +X/-X"}.Apply(ctx)
		},
	})
}
