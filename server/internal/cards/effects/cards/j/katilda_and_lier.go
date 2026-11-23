package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Katilda and Lier — Legendary Creature — Human {G}{W}{U}, 3/3:
//
//	"Whenever you cast a Human spell, target instant or sorcery card in your
//	 graveyard gains flashback until end of turn. The flashback cost is equal
//	 to its mana cost."
//
// Snapcaster Mage's grant (ADR 0066) behind a cast trigger: the permission
// names that card object, leaves its cost at the card's mana cost, and
// carries CR 702.34a's exile replacement, so the card is cast once and
// exiled. The trigger fires on a Human spell, judged through the spell's
// effective creature types (a changeling qualifies, CR 702.73a); Katilda
// casting herself does not trigger it, since she is not on the battlefield
// as she is cast.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1990012e-2a92-4e76-917a-f90fb7a4ed66",
		Name:         "Katilda and Lier",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WheneverYouCast(effects.Subtype("Human"), "Katilda and Lier — target instant or sorcery card in your graveyard gains flashback until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					ts := ctx.LegalTargets()
					if len(ts) == 0 {
						return nil
					}
					return effects.GrantFlashbackToCard{
						Target: ts[0].ID,
						Label:  "Flashback — its mana cost (Katilda and Lier)",
					}.Apply(ctx)
				}),
			effects.TargetCardInGraveyard("target instant or sorcery card in your graveyard",
				effects.YouOwn(), effects.Or(effects.Instant(), effects.Sorcery())),
		)},
	})
}
