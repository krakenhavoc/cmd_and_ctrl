package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thorn of the Black Rose — Creature — Human Assassin {3}{B}, 1/3
// (Commander Masters):
//
//	"Deathtouch
//	 When this creature enters, you become the monarch."
//
// Deathtouch is printed data; the entry trigger is Archon of Coronation's
// (CR 725.1).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a7894a81-8c0a-4015-8bc1-2698cf5306d1",
		Name:            "Thorn of the Black Rose",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Thorn of the Black Rose — you become the monarch", effects.Do(effects.BecomeTheMonarch{})),
		},
	})
}
