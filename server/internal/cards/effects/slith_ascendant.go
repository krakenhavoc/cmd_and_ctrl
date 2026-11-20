package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Slith Ascendant — Creature — Slith {1}{W}{W}, 1/1 (Mirrodin, #5606):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, put a +1/+1
//	 counter on it."
//
// Printed flying, and Hexgold Slith's trigger (mrdB5SlithGrows): the
// counter goes on the creature only while it is still the object that
// dealt the damage (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "72f5095f-b4ba-45ba-82e5-9a22bddd544b",
		Name:            "Slith Ascendant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{mrdB5SlithGrows("Slith Ascendant")},
	})
}
