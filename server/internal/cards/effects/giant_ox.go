package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Ox — {1}{W} Creature — Ox 0/6:
//
//	"This creature crews Vehicles using its toughness rather than its
//	 power."
//
// CrewsUsingToughness for crew only (ADR 0186 Decision 4, Interface Ace's
// shape): the Ox counts as 6 toward a crew cost and its real power stays
// 0. A saddle cost still counts its plain power.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b42bdf54-9055-405c-87e6-0912dbe2bd68",
		Name:         "Giant Ox",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			CrewsUsingToughness(game.CostPowerCrew),
		},
	})
}
