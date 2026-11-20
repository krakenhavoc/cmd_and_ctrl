package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skophos Reaver — Creature — Minotaur Warrior {2}{R}, 2/3 (Modern
// Horizons 2, #5501):
//
//	"During your turn, this creature gets +2/+0.
//	 Madness {1}{R} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your
//	 graveyard.)"
//
// Dapper Shieldmate's layer 7c static, on only while its controller is
// the active player, and the engine's madness (CR 702.35a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4bc0bdd5-32da-46df-b624-29a709efc676",
		Name:         "Skophos Reaver",
		Completeness: CompletenessFull,
		Madness:      "{1}{R}",
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
