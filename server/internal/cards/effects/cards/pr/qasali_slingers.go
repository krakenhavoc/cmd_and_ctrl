package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Qasali Slingers — Creature — Cat Warrior {4}{G}, 3/5 (Commander 2017,
// #5533):
//
//	"Reach
//	 Whenever this creature or another Cat you control enters, you may
//	 destroy target artifact or enchantment."
//
// Aura Shards' optional Naturalize trigger on Kemba, Kha Enduring's
// condition: the Slingers themselves, or any Cat entering under their
// controller's control (a token Cat too; the subtype a word of the text,
// CR 612.2). The target is chosen as the trigger goes on the stack and the
// "you may" is asked as it resolves (CR 603.5); a target gone by then
// leaves nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f763d118-3183-4604-a50c-7c3933e2185e",
		Name:            "Qasali Slingers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.KembaOrAnotherCatEntered,
			Targets:   effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
			Key:       "Qasali Slingers — destroy target artifact or enchantment",
			Effect:    effects.DestroyChosenPermanent,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Qasali Slingers — destroy an artifact or enchantment?",
			},
		}},
	})
}
