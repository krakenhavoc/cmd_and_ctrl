package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "f6d3f145-00e0-47cd-91d8-0381c19a000d",
		Name:            "Kaito, Cunning Infiltrator",
		Completeness:    CompletenessFull,
		StartingLoyalty: 3,
		Emblem: &EmblemSpec{
			Label: "Kaito, Cunning Infiltrator emblem",
			Text:  "Whenever a player casts a spell, you create a 2/1 blue Ninja creature token.",
			Triggered: []game.TriggeredAbility{
				On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Kind == game.EventCast && ev.Actor != uuid.Nil
				}, "Kaito, Cunning Infiltrator emblem — create a 2/1 blue Ninja creature token",
					Do(CreateToken{Template: TokenCard("2/1 blue Ninja"), N: 1})),
			},
		},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return combatDamageToPlayerBy(ev, source.Controller, g)
			},
			Key:    "Kaito, Cunning Infiltrator — put a loyalty counter on Kaito",
			Effect: stdB15LoyaltyCounterOnThis,
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "+1: Up to one target creature you control can't be blocked this turn. Draw a card, then discard a card.",
				Cost:    LoyaltyCost(1),
				Targets: UpToOneTargetCreature("up to one target creature you control", YouControl()),
				Purpose: kaitoPlusOnePurpose(),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind != game.TargetCard {
							continue
						}
						if err := (RestrictUntilEOT{
							Target:       t.ID,
							Restrictions: game.CantBeBlocked,
							Label:        "Kaito, Cunning Infiltrator — can't be blocked",
						}).Apply(ctx); err != nil {
							return err
						}
					}
					return lootOne(g, item, 1)
				},
			},
			{
				Label:  "−2: Create a 2/1 blue Ninja creature token.",
				Cost:   LoyaltyCost(-2),
				Effect: Do(CreateToken{Template: TokenCard("2/1 blue Ninja"), N: 1}),
			},
			{
				Label:  "−9: You get an emblem with \"Whenever a player casts a spell, you create a 2/1 blue Ninja creature token.\"",
				Cost:   LoyaltyCost(-9),
				Effect: Do(CreateEmblem{}),
			},
		},
	})
}

// kaitoPlusOnePurpose is the +1's declaration: the target can't be blocked
// this turn (a benefit to the creature it names), and the controller loots.
func kaitoPlusOnePurpose() game.Purpose {
	p := ForTargets(CantBeBlockedUntilEOT(0, 0, 0))
	p.Draws, p.Discards = 1, 1
	return p
}
