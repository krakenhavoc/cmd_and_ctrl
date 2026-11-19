package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Niblis of the Urn — Creature — Spirit {1}{W}, 1/1 (Innistrad Remastered):
//
//	"Flying
//	 Whenever this creature attacks, you may tap target creature."
//
// The target is chosen as the trigger is put on the stack; the "may" is
// asked on resolution, as Artisan of Kozilek's is.
func init() {
	attack := WheneverThisAttacks("Niblis of the Urn — you may tap target creature", tapChosenPermanent)
	attack.Targets = TargetCreature("target creature")
	Register(Spec{
		OracleID:        "3ea7a555-d6fa-4cc9-9348-784e6f9a8cfc",
		Name:            "Niblis of the Urn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Optional(attack, "Niblis of the Urn — tap target creature?"),
		},
	})
}
