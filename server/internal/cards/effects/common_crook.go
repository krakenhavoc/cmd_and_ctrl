package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Common Crook — Creature — Human Rogue Villain {1}{B}, 2/2 (Marvel Super
// Heroes Commander, #4455): "When this creature dies, create a Treasure
// token." A dies trigger (CR 700.4) that makes the catalog Treasure.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ebcbfdba-34c7-40eb-91a9-d306ff5ae8b9",
		Name:         "Common Crook",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Common Crook — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
