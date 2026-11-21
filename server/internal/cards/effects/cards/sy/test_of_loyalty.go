package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Test of Loyalty — Sorcery {B} (Star Trek, #4973):
//
//	"Choose one —
//	 • Plain, Simple Garak — Exile target card from a graveyard. Draw a
//	   card.
//	 • Of the Obsidian Order — Return target creature card from your
//	   graveyard to your hand."
//
// The mode names are flavor words with no rules meaning (CR 207.2d). Each
// bullet has one target; a bullet whose target is gone by resolution is
// the spell's only target, so the spell does nothing at all (CR 608.2b)
// and the Garak bullet draws no card. Otherwise the exile comes first and
// the draw after it, in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6cb1c3fb-033a-4de5-ad07-69d9ad0ca5ae",
		Name:         "Test of Loyalty",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Plain, Simple Garak — Exile target card from a graveyard. Draw a card.",
				effects.TargetCardInGraveyard("target card from a graveyard"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					target, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (effects.ExileTarget{Target: target.ID}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
				}), game.Purpose{Draws: 1}),
			effects.ModeDoing("Of the Obsidian Order — Return target creature card from your graveyard to your hand.",
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.YouOwn(), effects.Creature()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					card, found := effects.ModeTarget(ctx, occ)
					if !found {
						return nil
					}
					return effects.ReturnFromGraveyard{Target: card.ID, Dest: game.ZoneHand}.Apply(ctx)
				}),
		),
	})
}
