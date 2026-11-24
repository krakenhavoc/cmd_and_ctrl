package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horned Sliver — Creature — Sliver {2}{G}, 2/2 (Secret Lair Drop):
//
//	"All Sliver creatures have trample."
//
// A layer-6 keyword grant (CR 613.1f) of trample to every Sliver creature at
// the table, an opponent's and this one included (Battering Sliver's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64553181-9852-459f-a6e0-54ce188dd937",
		Name:         "Horned Sliver",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalKeywordGrant(effects.TribeFilter{Tribes: []string{"Sliver"}}, "trample"),
		},
	})
}
