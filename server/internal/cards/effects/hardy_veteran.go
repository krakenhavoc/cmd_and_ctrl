package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hardy Veteran — Creature — Human Warrior {1}{G}, 2/2 (Rivals of
// Ixalan, #5531):
//
//	"During your turn, this creature gets +0/+2."
//
// Dapper Shieldmate's turn-gated pump with the bonus on toughness: a layer
// 7c static re-read on every pass, on only while its controller is the
// active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fc112282-bc40-4b91-ac5f-ae49db1dabb8",
		Name:         "Hardy Veteran",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:    "During your turn, this creature gets +0/+2.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			},
			Apply: pumpBy(0, 2),
		}},
	})
}
