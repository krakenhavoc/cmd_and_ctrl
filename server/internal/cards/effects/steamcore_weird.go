package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:       "e8f6caf4-4c6c-468a-9388-88c323ffe30b",
		Name:           "Steamcore Weird",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"R"},
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(gptWhenEntersIfSpent("R", "Steamcore Weird — 2 damage to any target",
				TargetAny(), sourceDealsDamageToEachLegalTarget(2)), ForTargets(DamageToTarget(0, 2))),
		},
	})
}
