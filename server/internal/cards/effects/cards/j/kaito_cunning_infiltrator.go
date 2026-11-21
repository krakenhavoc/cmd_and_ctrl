package cardsj

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kaito, Cunning Infiltrator — Legendary Planeswalker — Kaito {1}{U}{U},
// starting loyalty 3 (Foundations, #4465):
//
//	"Whenever a creature you control deals combat damage to a player, put a
//	 loyalty counter on Kaito.
//	 +1: Up to one target creature you control can't be blocked this turn.
//	 Draw a card, then discard a card.
//	 −2: Create a 2/1 blue Ninja creature token.
//	 −9: You get an emblem with "Whenever a player casts a spell, you create
//	 a 2/1 blue Ninja creature token.""
//
// The trigger is Ikra Shidiqi's condition (one per creature that connects
// with a player); the counter goes on Kaito only while he is still on the
// battlefield. The +1 chooses no target or one: with a target that has become
// illegal the whole ability does nothing, the loot included (CR 608.2b),
// and with none it loots alone (Reconnaissance Mission's loot, Access
// Tunnel's restriction). The emblem is Chandra's shape: its trigger watches
// every cast, whoever casts, and its controller, the player who got the
// emblem, creates the Ninja (CR 114.4); it triggers before the spell
// resolves and whether or not the spell does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6d3f145-00e0-47cd-91d8-0381c19a000d",
		Name:            "Kaito, Cunning Infiltrator",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &effects.EmblemSpec{
			Label: "Kaito, Cunning Infiltrator emblem",
			Text:  "Whenever a player casts a spell, you create a 2/1 blue Ninja creature token.",
			Triggered: []game.TriggeredAbility{
				effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventCast && ev.Actor != uuid.Nil
				}, "Kaito, Cunning Infiltrator emblem — create a 2/1 blue Ninja creature token",
					effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 blue Ninja"), N: 1})),
			},
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.CombatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key:    "Kaito, Cunning Infiltrator — put a loyalty counter on Kaito",
			Effect: effects.StdB15LoyaltyCounterOnThis,
		}},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: Up to one target creature you control can't be blocked this turn. Draw a card, then discard a card.",
				Cost:    effects.LoyaltyCost(1),
				Targets: effects.UpToOneTargetCreature("up to one target creature you control", effects.YouControl()),
				Purpose: kaitoPlusOnePurpose(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (effects.RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Kaito, Cunning Infiltrator — can't be blocked",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.LootOne(g, item, 1)
				},
			},
			{
				Label:  "−2: Create a 2/1 blue Ninja creature token.",
				Cost:   effects.LoyaltyCost(-2),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("2/1 blue Ninja"), N: 1}),
			},
			{
				Label:  "−9: You get an emblem with \"Whenever a player casts a spell, you create a 2/1 blue Ninja creature token.\"",
				Cost:   effects.LoyaltyCost(-9),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// kaitoPlusOnePurpose is the +1's declaration: the target can't be blocked
// this turn (a benefit to the creature it names), and the controller loots.
func kaitoPlusOnePurpose() game.Purpose {
	p := effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0))
	p.Draws, p.Discards = 1, 1
	return p
}
