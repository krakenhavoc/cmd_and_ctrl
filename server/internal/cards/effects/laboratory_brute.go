package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Laboratory Brute — Creature — Zombie Horror {3}{U}, 3/3 (Eldritch Moon,
// #5541):
//
//	"When this creature enters, mill four cards."
//
// Wailing Ghoul's enters trigger with four cards.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dc92bbee-a445-4927-932b-15985bb93f19",
		Name:         "Laboratory Brute",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Laboratory Brute — mill four cards", Do(MillCards{N: 4})),
		},
	})
}
