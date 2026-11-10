package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Noggle Robber — Creature — Noggle Rogue {1}{R/G}{R/G} (Lorwyn Eclipsed, #4458):
//
//	"When this creature enters or dies, create a Treasure token."
func init() {
	Register(Spec{
		OracleID:     "a2f33166-6cfc-4dc5-8edc-088bf40f06b5",
		Name:         "Noggle Robber",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrDies("Noggle Robber — create a Treasure token", Do(CreateToken{Template: TreasureToken(), N: 1})),
		},
	})
}
