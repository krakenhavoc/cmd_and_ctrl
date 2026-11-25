package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Release to the Wind — Instant {2}{U} (Rivals of Ixalan, S82 OLD batch 218,
// #5531):
//
//	"Exile target nonland permanent. For as long as that card remains exiled,
//	 its owner may cast it without paying its mana cost."
//
// Aerial Extortionist's exile with a grant over the card for its OWNER (not
// the spell's controller), cast-only and lasting while it stays in exile (a
// card exiled again later by something else does not inherit it), at no mana
// cost. A token exiled this way is gone (CR 111.8). A target that is no longer
// legal as the spell resolves is not exiled (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49d6a9be-1d10-4ce4-afa3-aa637f39b898",
		Name:         "Release to the Wind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, found := ctx.Game.LookupCardForEffect(id)
			if !found {
				return nil
			}
			return ctx.Game.ExileCardWithPermissionForEffect(id, game.CastPermission{
				Player:     c.Owner,
				Duration:   game.WhileInZoneDuration(),
				CastOnly:   true,
				Cost:       "{0}",
				AltCostKey: game.AltCostKeyGrantedCost,
				Source:     ctx.Source(),
				SourceName: "Release to the Wind",
				Label:      "Release to the Wind — cast " + c.Name + " without paying its mana cost",
			})
		},
	})
}
