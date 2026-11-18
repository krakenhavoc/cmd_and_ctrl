package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dapper Shieldmate — Creature — Human Soldier {3}{W}, 2/2 (Streets of
// New Capenna):
//
//	"This creature enters with a shield counter on it. (If it would be
//	 dealt damage or destroyed, remove a shield counter from it
//	 instead.)
//	 During your turn, this creature gets +2/+0."
//
// The pump is a layer 7c static re-read on every pass: it is on only
// while its controller is the active player.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cdd795b-13b8-4812-89ee-d248368f915b",
		Name:         "Dapper Shieldmate",
		Completeness: CompletenessFull,
		Replacements: snc6Entry("Dapper Shieldmate"),
		Static: []game.StaticAbility{{
			Label:    "During your turn, this creature gets +2/+0.",
			Layer:    game.Layer7PT,
			SubLayer: game.SubLayer7C_Modify,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return thisPermanent(target, g, source) && isActivePlayer(g, source.Controller)
			},
			Apply: pumpBy(2, 0),
		}},
	})
}
