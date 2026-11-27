package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gloomwidow's Feast — Instant {3}{G} (Shadowmoor, #5591):
//
//	"Destroy target creature with flying. If that creature was blue or
//	 black, create a 1/2 green Spider creature token with reach. (It can
//	 block creatures with flying.)"
//
// The target is read through the layered view, so a creature that gains or
// loses flying is judged as it is (CR 608.2b). Its colours are read before
// it is destroyed, as it last existed (CR 608.2h). The printed text asks
// only whether it was blue or black, not whether it died, so the Feast's
// controller gets the Spider even when regeneration or indestructible
// saves the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ceae65e-938d-4e76-b3f6-7be70232549e",
		Name:         "Gloomwidow's Feast",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			blueOrBlack := false
			if c, found := ctx.Game.LookupCardForEffect(t.ID); found {
				if info, live := ctx.Game.PermanentForEffect(game.ObjectRef{ID: t.ID, Epoch: c.ObjectEpoch}); live {
					for _, col := range info.Characteristic.Colors {
						if col == "U" || col == "B" {
							blueOrBlack = true
						}
					}
				}
			}
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if !blueOrBlack {
				return nil
			}
			return effects.CreateToken{Template: effects.TokenCard("1/2 green Spider with reach"), N: 1}.Apply(ctx)
		},
	})
}
