package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Destroy the Evidence — Sorcery {4}{B} (Return to Ravnica, #5566):
//
//	"Destroy target land. Its controller reveals cards from the top of their
//	 library until they reveal a land card, then puts those cards into their
//	 graveyard."
//
// "Its controller" is read before the land is destroyed, from the battlefield
// object the spell targeted. The reveal-and-mill is Balustrade Spy's: the run is
// revealed, then the whole run, land included, goes to the graveyard. A land
// that survives (indestructible) still has its controller reveal. A library with
// no land reveals all of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "79aa86f5-bee4-498c-a0cd-719c05b136b2",
		Name:         "Destroy the Evidence",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			owner, ok := effects.ControllerOfTarget(ctx, id)
			if !ok {
				return nil
			}
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			return effects.OldB421RevealUntilLandAndMill(ctx, owner)
		},
	})
}
