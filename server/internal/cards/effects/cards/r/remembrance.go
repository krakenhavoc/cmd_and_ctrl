package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Remembrance — Enchantment {3}{W} (Urza's Saga, S82 OLD b722):
//
//	Whenever a nontoken creature you control dies, you may search your library
//	for a card with the same name as that creature, reveal it, put it into your
//	hand, then shuffle.
//
// A dies trigger for any nontoken creature you control (Midnight Reaper's
// condition); the name is the creature's as it last existed, and the search is
// the controller's "you may".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "02d0b0fd-ff73-4e26-be44-d85c743d51e4",
		Name:         "Remembrance",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventLTB, effects.S82OldB722YourNontokenCreatureDied,
				"Remembrance — you may search for a card with the same name", effects.S82OldB722SearchForTheSameName),
				game.Purpose{Tutors: 1, DeathPayoff: true}),
		},
	})
}
