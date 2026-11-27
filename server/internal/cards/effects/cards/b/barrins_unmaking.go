package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Barrin's Unmaking — Instant {1}{U} (Invasion):
//
//	"Return target permanent to its owner's hand if that permanent shares a
//	 color with the most common color among all permanents or a color tied for
//	 most common."
//
// #6207's SharesAMostCommonColor, read as the spell resolves (the clause is
// "if", CR 608.2c, not an intervening if): the colour count is over every
// permanent on the battlefield (layered colours, CR 105.2) and the target is
// judged as it stands then. A colourless target shares nothing and stays; a
// target that is no longer legal does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bd7ad427-11d9-40d1-b292-71f985711665",
		Name:         "Barrin's Unmaking",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target permanent"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			for _, t := range ctx.LegalTargets() {
				if t.Kind != game.TargetCard {
					continue
				}
				c, ok := ctx.Game.LookupCardForEffect(t.ID)
				if !ok || !effects.SharesAMostCommonColor(ctx.Game, &c) {
					return nil
				}
				return effects.BounceToHand{Target: t.ID}.Apply(ctx)
			}
			return nil
		},
	})
}
