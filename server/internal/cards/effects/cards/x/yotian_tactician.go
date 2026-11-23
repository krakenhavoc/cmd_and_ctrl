package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "0d9d7f04-6319-4596-8b74-de5a794936e5",
		Name:         "Yotian Tactician",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{Tribes: []string{"Soldier"}, Others: true, YoursOnly: true}, 1, 1),
		},
	})
}
