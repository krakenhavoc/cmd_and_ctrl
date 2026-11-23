package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desdemona, Freedom's Edge — Legendary Creature — Human Rogue {2}{R}{W},
// 3/4 (Fallout, #5948):
//
//	"Vigilance
//	 Whenever Desdemona attacks, target creature card in your graveyard that's
//	 an artifact or that has mana value 3 or less gains escape until end of
//	 turn. The escape cost is equal to its mana cost plus exile two other cards
//	 from your graveyard. (You may cast it from your graveyard for its escape
//	 cost this turn.)"
//
// Snapcaster Mage's single-card grant with escape's price (CR 702.138a):
// a cast permission over that one card, until end of turn, whose cost is the
// card's own mana cost (left empty) and exiling two other cards from the
// graveyard (ExileOtherFromGraveyard). Escape sets no ExileOnResolution: the
// card goes to the battlefield or the graveyard like any other. The target is
// a creature card in your graveyard that is an artifact or has mana value 3 or
// less, chosen as the trigger goes on the stack, and the grant is made only if
// it is still that card as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fd06d110-417c-42f5-a7b2-fd032f631dd2",
		Name:            "Desdemona, Freedom's Edge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WheneverThisAttacks("Desdemona — target creature card in your graveyard gains escape until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id, ok := effects.B16FirstLegalTargetCard(ctx)
					if !ok {
						return nil
					}
					g.GrantCastPermissionOverCardForEffect(id, game.CastPermission{
						Player:                  ctx.Controller(),
						Zone:                    game.ZoneGraveyard,
						AltCostKey:              "escape",
						ExileOtherFromGraveyard: 2,
						Source:                  ctx.Source(),
						Label:                   "Escape — its mana cost, Exile two other cards from your graveyard (Desdemona, Freedom's Edge)",
					})
					return nil
				}),
				effects.TargetCardInGraveyard("target creature card in your graveyard that's an artifact or that has mana value 3 or less",
					effects.Creature(), effects.YouOwn(), effects.Or(effects.Artifact(), effects.ManaValueLE(3)))),
		},
	})
}
