package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harrier Griffin — Creature — Griffin {5}{W}, 3/3 (Guildpact, #5599):
//
//	"Flying
//	 At the beginning of your upkeep, tap target creature."
//
// An upkeep trigger (AtYourUpkeep) with a creature target chosen as it
// goes on the stack (CR 603.3d), tapping it as it resolves
// (tapEachLegalTargetBody: nothing if it left in response). Any creature
// may be chosen, the Griffin included; with none on the battlefield the
// trigger is removed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "65dc3191-0eb2-4c10-bf2c-ee0e24888870",
		Name:            "Harrier Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(AtYourUpkeep("Harrier Griffin — tap target creature", tapEachLegalTargetBody),
				TargetCreature("target creature")),
		},
	})
}
