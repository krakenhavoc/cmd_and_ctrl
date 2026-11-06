package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// White Lotus Reinforcements — Creature — Human Soldier Ally {1}{G}{W},
// 2/3 (Avatar: The Last Airbender, #4459):
//
//	"Vigilance
//	 Other Allies you control get +1/+1."
//
// A lord: TribalAnthem over the other Ally creatures its controller
// controls, read through the layered view so a changeling counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c55a4b25-ccc4-49de-9f98-a829c547a64d",
		Name:            "White Lotus Reinforcements",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Ally"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
