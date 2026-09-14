package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "09955b4b-6052-4c27-8b63-9548482b3c5c",
		Name:         "Heritage Reclamation",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Destroy target artifact.", TargetPermanent("target artifact", Artifact())),
			Mode("Destroy target enchantment.", TargetPermanent("target enchantment", Enchantment())),
			Mode("Exile up to one target card from a graveyard. Draw a card.",
				TargetCardInGraveyard("up to one target card from a graveyard").WithCount(0, 1)),
		),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if ctx.HasMode(0) || ctx.HasMode(1) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (DestroyTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
			}
			if ctx.HasMode(2) {
				for _, t := range ctx.LegalTargets() {
					if t.Kind != game.TargetCard {
						continue
					}
					if err := (ExileTarget{Target: t.ID}).Apply(ctx); err != nil {
						return err
					}
				}
				return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
			}
			return nil
		},
	})
}
