package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bounty Agent — Creature — Human Soldier {1}{W}, 2/2 (Commander 2020, S82
// OLD, #5513):
//
//	"Vigilance
//	 {T}, Sacrifice this creature: Destroy target legendary permanent that's
//	 an artifact, creature, or enchantment."
//
// Barbarian Riftcutter's sacrifice-to-destroy with {T} added: the {T} is the
// symbol (CR 302.6, so summoning sickness applies) and the Agent is paid as a
// cost, so the ability resolves with its source gone. The target is a
// legendary artifact, creature or enchantment, judged again on resolution
// (CR 608.2b): a permanent that stopped being legendary, or left, is spared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "056b5d53-f2fd-4e83-9fd3-d8f0d0be64b9",
		Name:            "Bounty Agent",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: Destroy target legendary permanent that's an artifact, creature, or enchantment.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetPermanent("target legendary permanent that's an artifact, creature, or enchantment",
				effects.Legendary(), effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment())),
			Effect: effects.DestroyChosenPermanent,
		}},
	})
}
