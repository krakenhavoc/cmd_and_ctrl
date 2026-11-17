package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Yotian Tactician — Creature — Human Soldier {2}{W}{U}, 3/4 (The
// Brothers' War, #5482):
//
//	"Other Soldiers you control get +1/+1."
//
// A tribal anthem (Morcant's Loyalist's shape) over the other Soldiers
// the Tactician's controller controls.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0d9d7f04-6319-4596-8b74-de5a794936e5",
		Name:         "Yotian Tactician",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
