package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chasm Drake — Creature — Drake {4}{U}, 3/3 (Magic 2012, #5573):
//
//	"Flying
//	 Whenever this creature attacks, target creature you control gains
//	 flying until end of turn."
//
// Finch Formation's flying grant on an attack trigger instead of an enters
// one; the Drake may target itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0f44bb8-f07b-416c-b767-87b9fe6461aa",
		Name:            "Chasm Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WheneverThisAttacks("Chasm Drake — target creature you control gains flying until end of turn",
				ebTargetCreatureGainsUntilEOT("flying", "Chasm Drake — flying until end of turn")),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
