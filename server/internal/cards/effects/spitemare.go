package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spitemare — Creature — Elemental {2}{R/W}{R/W}, 3/3 (Eventide, #5590):
//
//	"Whenever this creature is dealt damage, it deals that much damage to
//	 any target."
//
// Boros Reckoner's trigger: one per damaged batch, with the total it was
// dealt as the amount, and the target chosen as it goes on the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf4d8b19-fd11-4200-9871-896772019f4c",
		Name:         "Spitemare",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(cmdr2B08EnrageTrigger("Spitemare — it deals that much damage to any target", damagedCreatureReflectsChosenAmount),
				TargetAny()),
		},
	})
}
