package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spreading Algae — Enchantment — Aura {G} (Eighth Edition, S82 older sets
// batch 617, #5607):
//
//	"Enchant Swamp
//	 When enchanted land becomes tapped, destroy it.
//	 When this Aura is put into a graveyard from the battlefield, return it to
//	 its owner's hand."
//
// Contaminated Ground's tap trigger with a destroy, and Glistening Oil's
// return from the graveyard (the card is a new object once it leaves, so a
// card that has moved on is not found, CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a08189f2-1643-496a-8a91-74864528b4db",
		Name:         "Spreading Algae",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantLand(effects.HasSubtype("Swamp")),
		Purpose:      game.Purpose{Aura: game.AuraHarm},
		Triggered: []game.TriggeredAbility{
			{
				Watches:   []game.EventKind{game.EventTapCard, game.EventAttack},
				AppliesTo: effects.EnchantedPermanentBecameTapped,
				Key:       "Spreading Algae — destroy enchanted land",
				Effect: func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					id := item.Trigger.Event.CardID
					if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
						return nil
					}
					return effects.DestroyTarget{Target: id}.Apply(effects.NewContext(g, item))
				},
			},
			effects.WhenThisDies("Spreading Algae — return it to its owner's hand", effects.ReturnSelfFromGraveyardToHand),
		},
	})
}
