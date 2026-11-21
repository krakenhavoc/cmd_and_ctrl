package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barging Sergeant — Creature — Minotaur Soldier {4}{R}, 4/2:
//
//	"Haste
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor (CR 702.134) is the shared Mentor() trigger; see
// hammer_dropper.go for what its source-relative clause guarantees.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "af422a4b-087c-4753-9288-c31fb3457740",
		Name:            "Barging Sergeant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.Mentor("Barging Sergeant — mentor")},
	})
}
