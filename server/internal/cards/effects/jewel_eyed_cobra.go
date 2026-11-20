package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jewel-Eyed Cobra — Creature — Snake {2}{G}, 3/1 (Modern Horizons 2,
// #5501):
//
//	"Deathtouch
//	 When this creature dies, create a Treasure token. (It's an artifact
//	 with "{T}, Sacrifice this token: Add one mana of any color.")"
//
// Dire Fleet Hoarder's dies trigger on a deathtouch body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bbfee53c-30a6-4ed0-beef-5c254be75fc5",
		Name:            "Jewel-Eyed Cobra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Jewel-Eyed Cobra — create a Treasure token",
				Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
