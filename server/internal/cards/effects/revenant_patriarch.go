package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Revenant Patriarch — Creature — Spirit {4}{B}, 4/3 (Guildpact, #5599):
//
//	"When this creature enters, if {W} was spent to cast it, target
//	 player skips their next combat phase.
//	 This creature can't block."
//
// Gruul Scrapper's intervening if (gptWhenEntersIfSpent) on Blinding
// Angel's skip record (gptTargetSkipsNextCombatPhase): the target skips
// one combat phase, the next one of theirs that would begin, this turn's
// if it is still to come (CR 500.11, 614.10). With no white mana spent
// nothing triggers. The block restriction is Fallen Askari's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "942f6260-055e-4cc7-bec6-3a46fd4b048e",
		Name:           "Revenant Patriarch",
		Completeness:   CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"W"},
		Triggered: []game.TriggeredAbility{
			gptWhenEntersIfSpent("W", "Revenant Patriarch — target player skips their next combat phase",
				TargetPlayer("target player"), gptTargetSkipsNextCombatPhase),
		},
		Static: []game.StaticAbility{RestrictSelf(game.CantBlock)},
	})
}
