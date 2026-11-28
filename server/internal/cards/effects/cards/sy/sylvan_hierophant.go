package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sylvan Hierophant — Creature — Human Cleric {1}{G}, 1/2 (Weatherlight):
//
//	"When this creature dies, exile it, then return another target creature card
//	 from your graveyard to your hand."
//
// The target is chosen as the trigger goes on the stack and is never the
// Hierophant itself. The two instructions are sequential, not conditional: the
// card is exiled first (so it can't be returned by its own ability), and the
// return happens whether or not the exile did, so long as the target is still
// legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e98361d9-1e3a-464e-bac6-75314fc49f32",
		Name:         "Sylvan Hierophant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisDies("Sylvan Hierophant — exile it, then return another target creature card from your graveyard to your hand",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						if z := g.FindCardZoneForEffect(item.SourceCardID); z == nil || z.Kind != game.ZoneGraveyard {
							return effects.ReturnLegalGraveyardTargetsToHand(ctx)
						}
						return effects.ExileTarget{Target: item.SourceCardID, ThenTail: ctx.ItemTail(sylvanHierophantExiled)}.Apply(ctx)
					}),
				effects.Another(effects.TargetCardInGraveyard("another target creature card from your graveyard", effects.YouOwn(), effects.Creature()))),
		},
	})
}

// sylvanHierophantExiled is the clause after the exile. An on-disk key.
var sylvanHierophantExiled = effects.CardTail("s82-old-b775/sylvan-hierophant-after-exile", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.ReturnLegalGraveyardTargetsToHand(ctx)
})
