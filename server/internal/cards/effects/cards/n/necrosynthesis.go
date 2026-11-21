package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

const necrosynthesisGrant = "necrosynthesis/another-creature-dies-counter" // an on-disk identity: never renamed

// Necrosynthesis — Enchantment — Aura {1}{B} (Innistrad: Midnight Hunt):
//
//	"Enchant creature
//	 Enchanted creature has "Whenever another creature dies, put a +1/+1
//	 counter on this creature."
//	 When enchanted creature dies, look at the top X cards of your
//	 library, where X is its power. Put one of those cards into your hand
//	 and the rest on the bottom of your library in a random order."
//
// The counter trigger is a granted bundle (ADR 0093), Elenda's "another
// creature dies" read for the creature that has it. The dies trigger is the
// Aura's own (WhenEnchantedCreatureDies' condition): X is the dead
// creature's last-known power (CR 603.10a, 608.2h), taken as the trigger is
// built so a later change cannot move it, and "your library" is the Aura's
// controller's. The look and the pick are Fomori Vault's
// TakeFromLibraryToHand: one card is taken (mandatory when there is one)
// and the rest go to the bottom in a random order.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "661618a2-54c7-4136-beff-fff5ff7d84cb",
		Name:         "Necrosynthesis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Grants: []effects.AbilityGrant{{
			Key: necrosynthesisGrant,
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventLTB, effects.AnotherCreatureDied, "Necrosynthesis — put a +1/+1 counter on this creature",
					effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			},
			Text: "Whenever another creature dies, put a +1/+1 counter on this creature.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(necrosynthesisGrant)},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.EquippedCreatureDied(ev, source)
			},
			Key: "Necrosynthesis — look at the top X cards of your library, where X is its power; put one into your hand and the rest on the bottom in a random order",
			Build: func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Necrosynthesis — look at the top X cards of your library, where X is its power; put one into your hand and the rest on the bottom in a random order")
				item.Params.Amount = effects.B13LastKnownPower(g, ev.CardID, lki)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				x := item.Params.Amount
				if x <= 0 {
					return nil
				}
				player := item.Controller
				return effects.TakeFromLibraryToHand{
					Player: player,
					Cards:  g.LookAtTopOfLibraryForEffect(player, x),
					Max:    1,
					Label:  "Necrosynthesis — put one of those cards into your hand",
					Key:    effects.TakeRestBottomRandom,
				}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
