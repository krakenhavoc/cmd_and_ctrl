package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aura Fracture — Enchantment {2}{W} (Prophecy, S82 OLD b672, #5617):
//
//	"Sacrifice a land: Destroy target enchantment."
//
// Baloth Prime's sacrifice-a-land cost over Barbarian Riftcutter's destroy: the
// land is paid at announce, and the target may be any enchantment, an Aura or
// the Fracture itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3495d83a-b103-42be-8708-9ce971b352bd",
		Name:         "Aura Fracture",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice a land: Destroy target enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.B08SacrificeALand(),
			Targets: effects.TargetPermanent("target enchantment", effects.Enchantment()),
			Effect:  effects.DestroyChosenPermanent,
		}},
	})
}
