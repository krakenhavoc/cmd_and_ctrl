package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dire Fleet Hoarder — Creature — Human Pirate {1}{B}, 2/1:
//
//	"When this creature dies, create a Treasure token."
func init() {
	Register(Spec{
		OracleID:     "7d48ad95-d428-4e2d-b697-56907c3ce579",
		Name:         "Dire Fleet Hoarder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Dire Fleet Hoarder — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
