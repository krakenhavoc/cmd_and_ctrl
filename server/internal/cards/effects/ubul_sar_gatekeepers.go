package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ubul Sar Gatekeepers — Creature — Zombie Soldier {3}{B}, 2/4:
//
//	"When this creature enters, if you control two or more Gates,
//	 target creature an opponent controls gets -2/-2 until end of turn."
//
// An intervening "if" (CR 603.4), Duergar Hedge-Mage's shape: the Gate
// count is checked as the creature enters (no trigger, and no target,
// otherwise) and again as the trigger resolves. A Gate is a land subtype,
// so a land that is a Gate by an effect counts.
//
// The target is chosen as the trigger goes on the stack (CR 603.3d); the
// -2/-2 ends at cleanup (CR 514.2), and it is declared as a harm so it is
// aimed at an opponent's creature, as on Black Dragon.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f4e15575-2277-4cb1-a343-e9407c050eb1",
		Name:         "Ubul Sar Gatekeepers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(InterveningIf(condControlTwoOrMoreOfSubtype("Gate"),
				On(game.EventETB, AllOf(Self, controlTwoOrMoreOfSubtype("Gate")),
					"Ubul Sar Gatekeepers — if you control two or more Gates, target creature an opponent controls gets -2/-2 until end of turn",
					boostTargetUntilEOT("Ubul Sar Gatekeepers — -2/-2 until end of turn", -2, -2))),
				TargetCreature("target creature an opponent controls", OpponentControls())),
				ForTargets(HarmUntilEOT(0, -2, -2))),
		},
	})
}
