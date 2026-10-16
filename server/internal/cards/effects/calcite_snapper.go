package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Calcite Snapper — Creature — Turtle {1}{U}{U}, 1/4:
//
//	"Shroud (This creature can't be the target of spells or abilities.)
//	 Landfall — Whenever a land you control enters, you may switch this
//	 creature's power and toughness until end of turn."
//
// The switch is layer 7d (CR 613.4d, #3996): it applies after every
// set, pump and counter, whatever their timestamps.
// The ability's source must still be the same object (CR 400.7). Its own switch does not target, so
// shroud does not stop it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36593200-e53c-46e4-a04e-8deec8273c65",
		Name:            "Calcite Snapper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"shroud"},
		Triggered: []game.TriggeredAbility{
			Optional(Landfall("Calcite Snapper — you may switch its power and toughness",
				switchThisPT("Calcite Snapper — switch its power and toughness")),
				"Switch Calcite Snapper's power and toughness until end of turn?"),
		},
	})
}
