package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "afe4a60c-747a-439b-bf61-cd323b8dffe7",
		Name:         "Clammy Prowler",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(
				effects.Targeting(effects.WheneverThisAttacks("Clammy Prowler — another target attacking creature can't be blocked this turn",
					effects.TargetCantBeBlockedThisTurn("Clammy Prowler")),
					effects.Another(effects.TargetCreature("another target attacking creature", effects.AttackingCreature()))),
				effects.ForTargets(effects.CantBeBlockedUntilEOT(0, 0, 0))),
		},
	})
}
