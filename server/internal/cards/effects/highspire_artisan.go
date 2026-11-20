package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Highspire Artisan — Creature — Elf Artificer {2}{G}, 0/3 (Kaladesh, #5539):
//
//	"Reach (This creature can block creatures with flying.)
//	 Fabricate 1 (When this creature enters, put a +1/+1 counter on it
//	 or create a 1/1 colorless Servo artifact creature token.)"
//
// The printed keyword is read by the engine. Fabricate is the shared keyword constructor (fabricate_keyword.go,
// CR 702.123a): the enters trigger asks counters or Servos as it
// resolves, and makes the Servos without asking once this creature has
// left the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fc8c54c4-a436-43cc-888b-f4ba3e751c16",
		Name:            "Highspire Artisan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered:       []game.TriggeredAbility{Fabricate("Highspire Artisan", 1)},
	})
}
