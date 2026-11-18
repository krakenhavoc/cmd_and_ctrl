package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Festival Crasher — Creature — Devil {1}{R}, 1/3 (Midnight Hunt, #5497):
//
//	"Whenever you cast an instant or sorcery spell, this creature gets +2/+0
//	 until end of turn."
//
// Ghitu Amplifier's cast-an-instant-or-sorcery pump.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e6ddea98-b23f-44de-a63d-5d5dfad60e71",
		Name:         "Festival Crasher",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverYouCast(Or(Instant(), Sorcery()),
				"Festival Crasher — gets +2/+0 until end of turn",
				thisGetsUntilEndOfTurn(2, 0, "Festival Crasher — +2/+0")),
		},
	})
}
