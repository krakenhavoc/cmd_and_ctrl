package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heritage Reclamation — Instant {1}{G} (EDHREC rank 2657):
//
//	"Choose one —
//	 • Destroy target artifact.
//	 • Destroy target enchantment.
//	 • Exile up to one target card from a graveyard. Draw a card."
//
// Naturalize with a graveyard-hate cantrip mode. Three modes, one
// chosen (CR 700.2), each carrying its own target clause: the two
// destroys are a single hard target, the third is a zero-or-one
// target over every graveyard at the table — chosen with no target
// it is simply "draw a card", and chosen with a card that left the
// graveyard in response the spell is countered by game rules and
// draws nothing (CR 608.2b), as printed. The draw comes after the
// exile in printed order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "09955b4b-6052-4c27-8b63-9548482b3c5c",
		Name:         "Heritage Reclamation",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.Mode("Destroy target artifact.", effects.TargetPermanent("target artifact", effects.Artifact())),
			effects.Mode("Destroy target enchantment.", effects.TargetPermanent("target enchantment", effects.Enchantment())),
			effects.Mode("Exile up to one target card from a graveyard. Draw a card.",
				effects.TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if ctx.HasMode(0) || ctx.HasMode(1) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(2) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (effects.ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return effects.DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
