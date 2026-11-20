package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "f763d118-3183-4604-a50c-7c3933e2185e",
		Name:            "Qasali Slingers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: kembaOrAnotherCatEntered,
			Targets:   TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
			Key:       "Qasali Slingers — destroy target artifact or enchantment",
			Effect:    destroyChosenPermanent,
			OptionalPrompt: &game.TriggerOptionalPrompt{
				Question: "Qasali Slingers — destroy an artifact or enchantment?",
			},
		}},
	})
}
