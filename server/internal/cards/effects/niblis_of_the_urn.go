package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Niblis of the Urn — Creature — Spirit {1}{W}, 1/1 (Dark Ascension, #5571):
//
//	"Flying
//	 Whenever this creature attacks, you may tap target creature."
//
// Niblis of the Mist's optional tap on an attack trigger: the target,
// often the creature that would block, is picked as the trigger goes on
// the stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3ea7a555-d6fa-4cc9-9348-784e6f9a8cfc",
		Name:            "Niblis of the Urn",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(WheneverThisAttacks("Niblis of the Urn — you may tap target creature", tapTheFirstLegalTarget),
					"Niblis of the Urn — tap target creature?"),
				TargetCreature("target creature"),
			),
		},
	})
}
