package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Clammy Prowler — Enchantment Creature — Horror {3}{U}, 2/5 (Duskmourn:
// House of Horror, #4466):
//
//	"Whenever this creature attacks, another target attacking creature
//	 can't be blocked this turn."
//
// Helitrooper's attack trigger with Zhuge Jin's body
// (targetCantBeBlockedThisTurn): "another" excludes the Prowler by object
// (Another), and "attacking" is read as the target is chosen and again
// as the trigger resolves. Attacking alone, it has no target and is
// removed (CR 603.3d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "afe4a60c-747a-439b-bf61-cd323b8dffe7",
		Name:         "Clammy Prowler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(
				Targeting(WheneverThisAttacks("Clammy Prowler — another target attacking creature can't be blocked this turn",
					targetCantBeBlockedThisTurn("Clammy Prowler")),
					Another(TargetCreature("another target attacking creature", AttackingCreature()))),
				ForTargets(CantBeBlockedUntilEOT(0, 0, 0))),
		},
	})
}
