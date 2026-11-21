package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Nerd Rage — Enchantment — Aura {2}{U} (Fallout):
//
//	"Enchant creature
//	 When this Aura enters, draw two cards.
//	 Enchanted creature has "You have no maximum hand size" and "Whenever
//	 this creature attacks, if you have ten or more cards in hand, it gets
//	 +10/+10 until end of turn.""
//
// "You" in both granted abilities is the enchanted creature's controller,
// who may not be the Aura's: the first is a maximum-hand-size static that
// reaches the controller of the permanent the Aura is attached to
// (HandSizeAttachedController), live while it is attached. The second is a
// trigger the Aura grants, with an intervening "if" (CR 603.4) checked as
// the creature attacks and again as the ability resolves, on that
// controller's hand.
//
// No simplification.
const nerdRageGrant = "nerd-rage/big-attack"

func nerdRageTenCards(g *game.Game, controller uuid.UUID) bool {
	p := g.PlayerByIDForEffect(controller)
	return p != nil && p.Hand != nil && p.Hand.Size() >= 10
}

func init() {
	effects.Register(effects.Spec{
		OracleID:     "4601e4c7-74bc-44ce-9274-68ba79f8aae1",
		Name:         "Nerd Rage",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Draws: 2, Aura: game.AuraBenefit},
		HandSize:     []game.HandSizeStatic{{Players: game.HandSizeAttachedController, Kind: game.HandSizeNoMaximum}},
		Grants: []effects.AbilityGrant{{
			Key: nerdRageGrant,
			Triggered: []game.TriggeredAbility{
				effects.InterveningIf(effects.CondBoard{Name: "you have ten or more cards in hand", Test: func(ctx *effects.Context) bool {
					return nerdRageTenCards(ctx.Game, ctx.Controller())
				}}, effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.CardID == source.InstanceID && nerdRageTenCards(g, source.Controller)
				}, "Nerd Rage — if you have ten or more cards in hand, this creature gets +10/+10 until end of turn",
					func(g *game.Game, item *game.StackItem) error {
						return effects.BoostUntilEOT{Target: item.SourceCardID, Power: 10, Toughness: 10, Label: "Nerd Rage — +10/+10"}.Apply(effects.NewContext(g, item))
					})),
			},
			Text: "Whenever this creature attacks, if you have ten or more cards in hand, it gets +10/+10 until end of turn.",
		}},
		Static: []game.StaticAbility{effects.GrantAbilitiesToAttached(nerdRageGrant)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Nerd Rage — draw two cards", effects.Do(effects.DrawCards{N: 2})),
		},
	})
}
