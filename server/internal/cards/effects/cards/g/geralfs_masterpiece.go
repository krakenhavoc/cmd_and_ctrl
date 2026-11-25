package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Geralf's Masterpiece — Creature — Zombie Horror {3}{U}{U}, 7/7 (Shadows
// over Innistrad, #5543):
//
//	"Flying
//	 This creature gets -1/-1 for each card in your hand.
//	 {3}{U}, Discard three cards: Return this card from your graveyard to
//	 the battlefield tapped."
//
// The shrink is a layer 7c modification that reads its controller's hand
// live, with the hand-size hint so the layer cache is dropped whenever a
// hand changes (Nyxathid's shape, with "your" in place of a chosen player).
// The graveyard ability is Stitchwing Skaab's with a discard-three cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "dacab848-89ca-46cc-b796-21fc2c218e78",
		Name:            "Geralf's Masterpiece",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:             "This creature gets -1/-1 for each card in your hand.",
			Layer:             game.Layer7PT,
			SubLayer:          game.SubLayer7C_Modify,
			DependsOnHandSize: true,
			AppliesTo: func(target *game.Card, _ *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && source.Controller != uuid.Nil
			},
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				n := effects.CardsInHand(g, source.Controller)
				c.Power -= n
				c.Toughness -= n
			},
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{U}, Discard three cards: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}{U}"), effects.DiscardN(3, "three cards")),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromGraveyardTapped,
		}},
	})
}
