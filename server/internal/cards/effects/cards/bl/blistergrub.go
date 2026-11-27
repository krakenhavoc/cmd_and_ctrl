package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blistergrub — Creature — Phyrexian Horror {2}{B}, 2/2 (Scars of Mirrodin,
// S82 OLD batch 472, #5577):
//
//	"Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)
//	 When this creature dies, each opponent loses 2 life."
//
// Swampwalk is printed (Moor Fiend's), and the dies trigger is life loss, not
// damage, to each opponent still in the game.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f5580a77-a69b-4ee8-9809-4523097cd7ac",
		Name:            "Blistergrub",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Blistergrub — each opponent loses 2 life", effects.Do(effects.EachOpponentLosesLife{N: 2})),
		},
	})
}
