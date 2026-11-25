package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

const (
	drivenDespairDrivenGrant  = "driven-despair/driven-combat-damage-draw"
	drivenDespairDespairGrant = "driven-despair/despair-combat-damage-discard"
)

// Driven // Despair — Sorcery // Sorcery with aftermath (Hour of Devastation):
//
//	Driven {1}{G}: "Until end of turn, creatures you control gain trample and
//	 "Whenever this creature deals combat damage to a player, draw a card.""
//	Despair {1}{B}: "Aftermath (Cast this spell only from your graveyard.
//	 Then exile it.) Until end of turn, creatures you control gain menace and
//	 "Whenever this creature deals combat damage to a player, that player
//	 discards a card.""
//
// The creatures are the ones you control as the spell resolves (CR 611.2c).
// Each quoted trigger is a catalog bundle granted until end of turn (Predators'
// Hour's shape): each creature's own ability, so "you" is that creature's
// controller and "that player" is the player it dealt damage to.
func init() {
	const oracle = "87fca901-678d-4c9a-be4a-9dce2a53074b"
	yours := effects.And(effects.Creature(), effects.YouControl())
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Driven",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: drivenDespairDrivenGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Driven — draw a card", effects.Do(effects.DrawCards{N: 1})),
			},
			Text: "Whenever this creature deals combat damage to a player, draw a card.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"trample"}, Label: "Driven — trample"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{
				Match: yours,
				Keys:  []string{drivenDespairDrivenGrant},
				Label: "Driven — whenever this creature deals combat damage to a player, draw a card",
			}.Apply(ctx)
		},
	})
	effects.Register(effects.Spec{
		OracleID:     oracle + "#1",
		Name:         "Despair",
		Completeness: effects.CompletenessFull,
		Grants: []effects.AbilityGrant{{
			Key: drivenDespairDespairGrant,
			Triggered: []game.TriggeredAbility{
				effects.WheneverThisDealsCombatDamageToAPlayer("Despair — that player discards a card",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil {
							return nil
						}
						return effects.DiscardOne(g, item, item.Trigger.Event.Target)
					}),
			},
			Text: "Whenever this creature deals combat damage to a player, that player discards a card.",
		}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.GrantKeywordUntilEOT{Match: yours, Keywords: []string{"menace"}, Label: "Despair — menace"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantAbilitiesFor{
				Match: yours,
				Keys:  []string{drivenDespairDespairGrant},
				Label: "Despair — whenever this creature deals combat damage to a player, that player discards a card",
			}.Apply(ctx)
		},
	})
}
