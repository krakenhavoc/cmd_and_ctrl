package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vizier of the True — Creature — Human Cleric {3}{W}, 3/2:
//
//	"You may exert this creature as it attacks. (It won't untap during
//	 your next untap step.)
//	 Whenever you exert a creature, tap target creature an opponent
//	 controls."
//
// ADR 0130 §11: exert as it attacks (CR 701.43d) with no linked
// trigger, and a targeted "whenever you exert a creature" payoff that
// sees this creature and any other creature its controller exerts (the
// Amonkhet ruling). The target is chosen as the trigger goes on the
// stack and re-checked as it resolves (CR 608.2b). With no creature an
// opponent controls the trigger is removed (CR 603.3d).
//
// No simplification.
func init() {
	const label = "Vizier of the True — tap target creature an opponent controls"
	Register(Spec{
		OracleID:      "0e31d70e-2b5e-4c32-81cf-aed2a13e8aa5",
		Name:          "Vizier of the True",
		Completeness:  CompletenessFull,
		ExertOnAttack: ExertAsItAttacks(),
		Triggered: []game.TriggeredAbility{
			Targeting(
				WheneverYouExert(label, b36TapChosenCreature),
				TargetCreature("target creature an opponent controls", OpponentControls())),
		},
	})
}
