package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vengeful Rebel — Creature — Aetherborn Warrior {2}{B}, 3/2 (Aether Revolt,
// #5536):
//
//	"Revolt — When this creature enters, if a permanent left the
//	 battlefield under your control this turn, target creature an opponent
//	 controls gets -3/-3 until end of turn."
//
// Black Dragon's -3/-3 enters trigger behind the revolt intervening
// "if" (revolt.go, CR 603.4).
//
// No simplification.
func init() {
	ability := WhenThisEntersIfRevolt("Vengeful Rebel — target creature an opponent controls gets -3/-3 until end of turn",
		boostTargetUntilEOT("Vengeful Rebel — -3/-3 until end of turn", -3, -3))
	ability.Targets = TargetCreature("target creature an opponent controls", OpponentControls())
	Register(Spec{
		OracleID:     "273bf1a8-42a9-4db5-bd83-06db4cb5e5df",
		Name:         "Vengeful Rebel",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{TriggerWithPurpose(ability, ForTargets(HarmUntilEOT(0, -3, -3)))},
	})
}
