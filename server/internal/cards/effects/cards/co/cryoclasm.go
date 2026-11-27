package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cryoclasm — Sorcery {2}{R} (Coldsnap, #5597):
//
//	"Destroy target Plains or Island. Cryoclasm deals 3 damage to that land's
//	 controller."
//
// The land's controller is read before it is destroyed (CR 608.2h), so the
// damage lands on them even if the destroy does nothing (an indestructible
// land) and even though the land is gone by then. A target that is no longer
// legal as the spell resolves is not destroyed and its controller is not
// damaged: the spell is countered on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77c14618-ced1-4a46-a7fd-06cb2f7f38ff",
		Name:         "Cryoclasm",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target Plains or Island", effects.AnySubtype("Plains", "Island")),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			var owner game.Card
			found := false
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				if c, ok := ctx.Game.LookupCardForEffect(t.ID); ok {
					owner, found = c, true
				}
				break
			}
			if !found {
				return nil
			}
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			return effects.DealDamage{Source: ctx.Source(), Target: owner.Controller, Amount: 3}.Apply(ctx)
		},
	})
}
