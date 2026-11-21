package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vulpine Harvester — Creature — Phyrexian Fox {3}{W}, 3/3 (March of the
// Machine Commander):
//
//	"Whenever one or more Phyrexians you control attack, return target
//	 artifact card from your graveyard to the battlefield if its mana
//	 value is less than or equal to their total power."
//
// A batched attack trigger (CR 603.2c). The target is chosen as it goes on
// the stack; the total power of the Phyrexians attacking is fixed as the
// trigger is created (the Build stamps it on the item) and compared with
// the card's mana value as the ability resolves, so a Phyrexian that is
// removed from combat in response does not change "their" power. A card
// whose mana value is higher is simply not returned.
//
// No simplification.
func init() {
	const label = "Vulpine Harvester — return target artifact card from your graveyard to the battlefield if its mana value is at most the attackers' total power"
	trigger := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventAttack},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			if !effects.AttackDeclaredByYou(ev, source.Controller) {
				return false
			}
			c, ok := g.LookupCardForEffect(ev.CardID)
			return ok && c.IsCreature() && c.HasSubtype(source.TextSubtype("Phyrexian"))
		},
		Key: label,
		Build: func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			item := game.NewTriggeredItem(source, label)
			item.Params.Amount = effects.S82MomB11AttackingPhyrexiansPower(g, source.Controller, source.TextSubtype("Phyrexian"))
			return item
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			c, ok := g.LookupCardForEffect(ids[0])
			if !ok || c.ManaValue() > item.Params.Amount {
				return nil
			}
			return effects.ReturnFromGraveyardTogether{Targets: []uuid.UUID{ids[0]}}.Apply(ctx)
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "cc5e3a21-4911-49f8-8e3c-ab23812c4004",
		Name:         "Vulpine Harvester",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.OncePerBatch(trigger),
				effects.TargetCardInGraveyard("target artifact card from your graveyard", effects.Artifact(), effects.YouOwn())),
		},
	})
}
