package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Skyclave Geopede — Creature — Insect {2}{R}, 3/1 (Zendikar Rising,
// #5508):
//
//	"Trample
//	 Landfall — Whenever a land you control enters, this creature gets
//	 +2/+2 until end of turn."
//
// Trample is printed; the landfall pump is Saurian Explorer's (an ability
// word, CR 207.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e00d601f-ac73-47a6-8920-779d71b482c9",
		Name:            "Skyclave Geopede",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Triggered: []game.TriggeredAbility{
			Landfall("Skyclave Geopede — +2/+2 until end of turn (landfall)",
				thisGetsUntilEndOfTurn(2, 2, "Skyclave Geopede — +2/+2 until end of turn")),
		},
	})
}
