package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Subterranean Scout — Creature — Goblin Scout {1}{R}, 2/1 (Magic Origins, #5549):
//
//	"When this creature enters, target creature with power 2 or less can't be blocked this turn."
//
// Spider-Man, Hometown Hero's entry: the power bound is checked as the
// trigger goes on the stack and again as it resolves; the restriction
// ends in the cleanup step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "19f66bfa-bdfc-430a-ae39-4535f0b28324",
		Name:         "Subterranean Scout",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(
					WhenThisEnters("Subterranean Scout — target creature with power 2 or less can't be blocked this turn",
						nemTargetRestrictedThisTurn(game.CantBeBlocked, "Subterranean Scout — can't be blocked")),
					TargetCreature("target creature with power 2 or less", PowerLE(2))),
				ForTargets(CantBeBlockedUntilEOT(0, 0, 0))),
		},
	})
}
