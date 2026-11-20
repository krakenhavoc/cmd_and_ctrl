package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mesa Lynx — Creature — Cat {1}{W}, 2/1 (Zendikar Rising, #5508):
//
//	"During turns other than yours, this creature gets +0/+2."
//
// Sporeback Wolf's layer 7c static (CR 613.4c) turned around: re-read on
// every pass, on only while its controller is not the active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "211add65-b63d-43ad-8d69-b0598e003fef",
		Name:         "Mesa Lynx",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "During turns other than yours, this creature gets +0/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && !isActivePlayer(g, source.Controller)
			},
			Apply: pumpBy(0, 2),
		}},
	})
}
