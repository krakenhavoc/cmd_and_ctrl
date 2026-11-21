package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eater of Virtue — Legendary Artifact — Equipment {1} (Kamigawa: Neon
// Dynasty, #5493):
//
//	"Whenever equipped creature dies, exile it.
//	 Equipped creature gets +2/+0.
//	 As long as a card exiled with Eater of Virtue has flying, equipped
//	 creature has flying. The same is true for first strike, double
//	 strike, deathtouch, haste, hexproof, indestructible, lifelink,
//	 menace, protection, reach, trample, and vigilance.
//	 Equip {1}"
//
// The trigger exiles the dead creature's card from the graveyard, linked to
// this Equipment object (CR 607.2a, Game.ExileLinkedForEffect), and does
// nothing if the card has already left the graveyard or is a token. The
// pump is Skullclamp's. The keyword sentence is the Wretched Bonemass's
// and Soulflayer's: a layer 6 grant to the equipped creature of each
// listed keyword some card still exiled with this Equipment has (each card
// as it sits in exile, with its printed keywords; every protection ability
// with its quality, and a hexproof variant such as "hexproof from white"
// as that variant), so it stops counting when the card leaves exile
// (DependsOnExile). A different Eater of Virtue, or this one after it
// has left and returned, is a different object with its own cards
// (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6f8cc3b-b8f1-4b1a-8768-eddcb0631d80",
		Name:         "Eater of Virtue",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 0),
			{
				Label:          "As long as a card exiled with Eater of Virtue has flying, equipped creature has flying. The same is true for first strike, double strike, deathtouch, haste, hexproof, indestructible, lifelink, menace, protection, reach, trample, and vigilance.",
				Layer:          game.Layer6Ability,
				AppliesTo:      effects.AttachedToSource,
				DependsOnExile: true,
				Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
					var materials []game.Card
					for _, m := range g.ExiledWithForEffect(game.PermissionCardRef{ID: source.InstanceID, Epoch: source.ObjectEpoch}) {
						materials = append(materials, *m)
					}
					for _, kw := range effects.WretchedBonemassKeywords(materials) {
						c.Abilities = game.AppendKeywordAbility(c.Abilities, kw)
					}
					for i := range materials {
						for _, tok := range game.HexproofFromTokens(&materials[i]) {
							c.Abilities = game.AppendKeywordAbility(c.Abilities, tok)
						}
					}
				},
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.EquippedCreatureDied(ev, source)
			}, "Eater of Virtue — exile the equipped creature", func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil {
					return nil
				}
				id := item.Trigger.Event.CardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneGraveyard {
					return nil
				}
				ref, ok := g.SourceObjectForEffect(item)
				if !ok {
					return g.ExileCardForEffect(id)
				}
				return g.ExileLinkedForEffect(ref, id)
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
}
