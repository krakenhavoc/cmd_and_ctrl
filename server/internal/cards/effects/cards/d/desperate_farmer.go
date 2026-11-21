package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Desperate Farmer // Depraved Harvester — {2}{B} Creature — Human Peasant
// 2/2 // Creature — Human Knight 4/3 (Innistrad: Crimson Vow):
//
//	Front: "Lifelink
//	        When another creature you control dies, transform this creature."
//	Back:  "Lifelink"
//
// "Another" excludes the Farmer's own death. The transform is TransformThis,
// which does nothing once the Farmer has turned over or left (CR 701.27f).
//
// No simplification.
func init() {
	const oracle = "60d591f6-6d84-45d2-9d78-a2ef07a3e403"
	effects.Register(effects.Spec{
		OracleID:        oracle,
		Name:            "Desperate Farmer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AllOf(effects.ACreatureYouControlDied, effects.AnotherCreatureDied),
				"Desperate Farmer — transform this creature", effects.Do(effects.TransformThis{})),
		},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Depraved Harvester",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
	})
}
