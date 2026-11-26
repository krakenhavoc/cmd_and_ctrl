package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Eldrazi Skyspawner — Creature — Eldrazi Drone {2}{U}, 2/1 (Battle for
// Zendikar, #5547):
//
//	"Devoid (This card has no color.)
//	 Flying
//	 When this creature enters, create a 1/1 colorless Eldrazi Scion creature
//	 token. It has "Sacrifice this token: Add {C}.""
//
// Scion Summoner's token trigger, with printed flying.
//
// No simplification.
func init() {
	const name = "Eldrazi Skyspawner"
	effects.Register(effects.Spec{
		OracleID:        "b5661289-a03c-40fc-a1e4-f602adb56f81",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid, "flying"},
		Triggered:       []game.TriggeredAbility{effects.OldB307WhenThisEntersCreateScions(name, 1)},
	})
}
