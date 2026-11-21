package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Phyrexian Ingester — Creature — Phyrexian Beast {6}{U}, 3/3 (Commander
// Legends: Battle for Baldur's Gate):
//
//	"Imprint — When this creature enters, you may exile target nontoken
//	 creature.
//	 This creature gets +X/+Y, where X is the exiled creature card's power and
//	 Y is its toughness."
//
// Duplicant's optional linked exile (the card is exiled "with" this object,
// CR 607.2a), and a layer 7c boost that reads the last creature card exiled
// with the Ingester, as it sits in exile, so the printed power and toughness
// (a card off the battlefield has no layers). A card that leaves exile or an
// Ingester that is a new object ends the link and the boost (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec9c35f2-6a68-45e7-a86e-9f5af9fe28fc",
		Name:         "Phyrexian Ingester",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:          "This creature gets +X/+Y, where X is the exiled creature card's power and Y is its toughness.",
			Layer:          game.Layer7PT,
			SubLayer:       game.SubLayer7C_Modify,
			AppliesTo:      effects.SelfOnly,
			DependsOnExile: true,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				if exiled, ok := effects.B27LastCreatureCardExiledWith(g, source); ok {
					c.Power += exiled.Power
					c.Toughness += exiled.Toughness
				}
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches:        []game.EventKind{game.EventETB},
			AppliesTo:      effects.B06SelfETB,
			OptionalPrompt: &game.TriggerOptionalPrompt{Question: "Phyrexian Ingester — exile target nontoken creature?"},
			Targets:        effects.TargetCreature("target nontoken creature", effects.Not(effects.IsTokenPredicate())),
			Key:            "Phyrexian Ingester — you may exile target nontoken creature",
			Effect:         effects.ExileLinkedTargetsEffect,
		}},
	})
}
