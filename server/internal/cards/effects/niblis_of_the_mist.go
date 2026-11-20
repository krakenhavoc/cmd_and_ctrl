package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Niblis of the Mist — Creature — Spirit {2}{W}, 2/1 (Dark Ascension, #5571):
//
//	"Flying
//	 When this creature enters, you may tap target creature."
//
// Kitsune Dawnblade's optional enters trigger on a flyer: the target is
// picked as the trigger goes on the stack, and the tap is asked as it
// resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e059122d-b8d6-4878-bebb-e20ce6311d68",
		Name:            "Niblis of the Mist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(WhenThisEnters("Niblis of the Mist — you may tap target creature", tapTheFirstLegalTarget),
					"Niblis of the Mist — tap target creature?"),
				TargetCreature("target creature"),
			),
		},
	})
}
