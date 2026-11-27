package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Iridescent Drake — Creature — Drake {3}{U}, 2/2 (Urza's Destiny, S82 #5621):
//
//	"Flying
//	 When this creature enters, put target Aura card from a graveyard onto the
//	 battlefield under your control attached to this creature."
//
// Boonweaver Giant's enters-attached placement over a graveyard target (Mantle of
// the Ancients' ReturnFromGraveyardTogether with AttachedTo): the Aura enters
// under the Drake controller's control attached to the Drake, so CR 303.4f asks
// nothing. An Aura that can't legally enchant the Drake stays in its graveyard
// (CR 303.4i), and nothing is put onto the battlefield once the Drake has left
// (no object to attach to), or if the target has left the graveyard
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f8cf31db-6c33-4bfb-a24c-f22644ccd377",
		Name:            "Iridescent Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenThisEnters("Iridescent Drake — put target Aura card from a graveyard onto the battlefield under your control attached to it",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ids := effects.LegalTargetCardIDs(ctx)
					if len(ids) == 0 || !effects.OnBattlefield(g, item.SourceCardID) {
						return nil
					}
					return effects.ReturnFromGraveyardTogether{
						Targets:    ids,
						Controller: ctx.Controller(),
						AttachedTo: game.TargetRef{Kind: game.TargetCard, ID: item.SourceCardID},
					}.Apply(ctx)
				}),
				effects.TargetCardInGraveyard("target Aura card from a graveyard", effects.HasSubtype("Aura"))),
		},
	})
}
