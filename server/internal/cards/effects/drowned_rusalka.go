package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "62e86e38-2f8d-4cee-aa7e-4519add79d4d",
		Name:         "Drowned Rusalka",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{U}, Sacrifice a creature: Discard a card, then draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
			Cost:    Plus(ManaCost("{U}"), SacrificeACreature()),
			Effect:  s82BroB08Rummage,
		}},
	})
}
