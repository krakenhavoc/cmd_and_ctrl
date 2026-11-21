package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drowned Rusalka — Creature — Spirit {U}, 1/1 (Guildpact, #5599):
//
//	"{U}, Sacrifice a creature: Discard a card, then draw a card."
//
// Farid's rummage (s82BroB08Rummage, Charming Scoundrel's keyed discard
// then draw) behind High Market's sacrifice: the discard is chosen first
// and the draw follows it, drawing even from an empty hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "62e86e38-2f8d-4cee-aa7e-4519add79d4d",
		Name:         "Drowned Rusalka",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, Sacrifice a creature: Discard a card, then draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    effects.Plus(effects.ManaCost("{U}"), effects.SacrificeACreature()),
			Effect:  effects.S82BroB08Rummage,
		}},
	})
}
