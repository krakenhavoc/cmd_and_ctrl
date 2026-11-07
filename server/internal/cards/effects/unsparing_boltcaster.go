package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Unsparing Boltcaster — Creature — Ogre Wizard {2}{R}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"When this creature enters, it deals 5 damage to target creature an
//	 opponent controls that was dealt damage this turn."
//
// The target clause is Bitter Downfall's DealtDamageThisTurn narrowed to
// an opponent's creature: with no such creature the trigger has no legal
// target and is removed from the stack. The target is checked again as
// the trigger resolves (CR 608.2b), and the damage is dealt by the
// Boltcaster, as it last existed if it has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0a56a33b-033f-4398-8dc1-e531be059710",
		Name:         "Unsparing Boltcaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(WhenThisEnters("Unsparing Boltcaster — 5 damage to target creature an opponent controls that was dealt damage this turn",
				sourceDealsDamageToEachLegalTarget(5)),
				TargetCreature("target creature an opponent controls that was dealt damage this turn", OpponentControls(), DealtDamageThisTurn())),
				ForTargets(DamageToTarget(0, 5))),
		},
	})
}
