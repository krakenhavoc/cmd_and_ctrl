package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wanted Griffin — Creature — Griffin {3}{W}, 3/2 (Outlaws of Thunder
// Junction, #4468):
//
//	"Flying
//	 When this creature dies, create a 1/1 red Mercenary creature token
//	 with '{T}: Target creature you control gets +1/+0 until end of turn.
//	 Activate only as a sorcery.'"
//
// A dies trigger (CR 700.4) whose effect is the set's Mercenary template,
// made for the Griffin's controller as it died.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c1348afe-4dc7-41bb-9d3f-1e8751abd7db",
		Name:            "Wanted Griffin",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Wanted Griffin — create a 1/1 red Mercenary", createAMercenary()),
		},
	})
}
