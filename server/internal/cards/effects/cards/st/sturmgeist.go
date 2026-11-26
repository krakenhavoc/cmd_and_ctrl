package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sturmgeist — Creature — Spirit {3}{U}{U}, */* (Innistrad, S82 older sets
// batch 449, #5572):
//
//	"Flying
//	 Sturmgeist's power and toughness are each equal to the number of cards
//	 in your hand.
//	 Whenever this creature deals combat damage to a player, draw a card."
//
// Maro's characteristic-defining ability and a combat-damage draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b3045ee5-232b-4e63-8ea8-b5e46b49e71a",
		Name:            "Sturmgeist",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{
			effects.S82OldB021CardsInHandPT("Sturmgeist's power and toughness are each equal to the number of cards in your hand."),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Sturmgeist — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
