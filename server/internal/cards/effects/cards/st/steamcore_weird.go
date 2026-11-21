package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Steamcore Weird — Creature — Weird {3}{U}, 1/3 (Guildpact, #5599):
//
//	"When this creature enters, if {R} was spent to cast it, it deals 2
//	 damage to any target."
//
// Gruul Scrapper's intervening if (gptWhenEntersIfSpent) on Ghitu
// Slinger's damage trigger: with no red mana spent, or a Weird put onto
// the battlefield without being cast, nothing triggers. ReadsManaSpent
// asks what paid for a waived payment so the record is exact, and
// WantsColors steers the auto-tapper toward a red source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "e8f6caf4-4c6c-468a-9388-88c323ffe30b",
		Name:           "Steamcore Weird",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"R"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.GptWhenEntersIfSpent("R", "Steamcore Weird — 2 damage to any target",
				effects.TargetAny(), effects.SourceDealsDamageToEachLegalTarget(2)), effects.ForTargets(effects.DamageToTarget(0, 2))),
		},
	})
}
