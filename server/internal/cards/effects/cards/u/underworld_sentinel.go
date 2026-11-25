package cardsu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Underworld Sentinel — Creature — Skeleton Soldier {3}{B}{B}, 4/5:
//
//	"Whenever this creature attacks, exile target creature card from your
//	 graveyard.
//	 When this creature dies, put all cards exiled with it onto the
//	 battlefield."
//
// The exile is linked to the Sentinel as an object (CR 607.2a), read back
// through the dying Sentinel's last-known object when it dies. The cards
// enter under their owners' control, as the text does not say otherwise.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "eef75645-1aa5-4085-8cbd-8b6fec2cbdc2",
		Name:         "Underworld Sentinel",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WheneverThisAttacks("Underworld Sentinel — exile target creature card from your graveyard",
					func(g *game.Game, item *game.StackItem) error {
						ctx := effects.NewContext(g, item)
						link, ok := ctx.SourceRef()
						if !ok {
							return nil
						}
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return g.ExileLinkedKeyedForEffect(link, []uuid.UUID{id}, game.KeyedTail{})
					}),
				effects.TargetCardInGraveyard("target creature card from your graveyard", effects.Creature(), effects.YouOwn()),
			),
			effects.WhenThisDies("Underworld Sentinel — put all cards exiled with it onto the battlefield",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					link, ok := ctx.SourceRef()
					if !ok {
						return nil
					}
					var ids []uuid.UUID
					for _, c := range g.ExiledWithForEffect(game.PermissionCardRef(link)) {
						if !c.IsToken() {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil
					}
					return effects.PutCardsFromZoneTogetherWith(ctx, game.ZoneExile, ids, game.ZoneEntryOptions{}, game.TailRef{}, 0)
				}),
		},
	})
}
