package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret's Touch — {1}{U}{B} Enchantment — Aura (Aether Revolt, #5536):
//
//	"Enchant artifact
//	 Enchanted artifact is a creature with base power and toughness 5/5 in
//	 addition to its other types.
//	 When enchanted artifact is put into a graveyard, return that card to its
//	 owner's hand."
//
// A layer 4 type addition and a layer 7b base P/T. The return reads the card
// the event names and brings it back only if it is still in the graveyard
// (a token has nothing to return).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c6f2c767-3c8e-4113-b271-ea7c19583e6c",
		Name:         "Tezzeret's Touch",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantPermanent(effects.Artifact()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			{
				Label:     "Enchanted artifact is a creature in addition to its other types.",
				Layer:     game.Layer4Type,
				AppliesTo: effects.AttachedToSource,
				Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
					if !effects.EotHasType(c.Types, "Creature") {
						c.Types = append(append([]string(nil), c.Types...), "Creature")
					}
				},
			},
			effects.SetAttachedBasePT(5, 5),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenEnchantedCreatureDies("Tezzeret's Touch — return that card to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					return effects.ReturnFromGraveyard{Target: item.Trigger.Event.CardID, Dest: game.ZoneHand}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
