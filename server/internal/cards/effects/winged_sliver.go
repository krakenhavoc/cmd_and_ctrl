package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Winged Sliver — Creature — Sliver {1}{U}, 1/1 (Commander Legends: Battle
// for Baldur's Gate):
//
//	"All Sliver creatures have flying."
//
// Firewake Sliver's haste grant with flying, to every Sliver creature at the
// table in layer 6 (CR 613.1f), this one and an opponent's included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "24e96682-14ca-4a62-a920-2d686c8a5dbf",
		Name:         "Winged Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "flying"),
		},
	})
}
