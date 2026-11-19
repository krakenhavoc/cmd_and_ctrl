package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eldrazi Repurposer — Creature — Eldrazi Drone {2}{G}, 3/3 (Modern Horizons 3):
//
//	"Devoid
//	 When you cast this spell and when this creature dies, create a 0/1
//	 colorless Eldrazi Spawn creature token with "Sacrifice this token: Add
//	 {C}.""
func init() {
	const label = "Eldrazi Repurposer — create a 0/1 colorless Eldrazi Spawn creature token"
	Register(Spec{
		OracleID:        "ae3e5c80-6a6a-4781-a7b9-4d63357ffa3b",
		Name:            "Eldrazi Repurposer",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Tokens: 1, CreatureTokens: 1},
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{
			WhenYouCastThisSpell(label, gapB51EldraziSpawn),
			WhenThisDies(label, gapB51EldraziSpawn),
		},
	})
}
