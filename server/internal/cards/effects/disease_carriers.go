package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disease Carriers — Creature — Rat {2}{B}{B}, 2/2 (Urza's Destiny,
// #5621):
//
//	"When this creature dies, target creature gets -2/-2 until end of
//	 turn."
//
// A targeted dies trigger with the shared boostTargetUntilEOT body. The
// target is any player's creature, chosen as the trigger goes on the
// stack and checked again as it resolves (CR 608.2b); the -2/-2 ends in
// the cleanup step (CR 514.2). The target entry declares the harm, so
// the bot aims it at an opponent's creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ccdd6376-9b74-4d13-9f80-87d0bd3d4a2a",
		Name:         "Disease Carriers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			TriggerWithPurpose(Targeting(
				WhenThisDies("Disease Carriers — target creature gets -2/-2 until end of turn",
					boostTargetUntilEOT("Disease Carriers — -2/-2 until end of turn", -2, -2)),
				TargetCreature("target creature"),
			), ForTargets(HarmUntilEOT(0, -2, -2))),
		},
	})
}
