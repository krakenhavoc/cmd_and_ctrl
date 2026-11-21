package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Orzhov Enforcer — Creature — Human Rogue {1}{B}, 1/2 (Ravnica Allegiance,
// #5521):
//
//	"Deathtouch
//	 Afterlife 1 (When this creature dies, create a 1/1 white and black
//	 Spirit creature token with flying.)"
//
// Deathtouch is printed data (CR 702.2b). Afterlife is the dies trigger it
// reminds (CR 702.135a), Syndicate Messenger's 1/1 white and black Spirit
// with flying.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f60b879e-82b4-442d-a544-2349429fe4c9",
		Name:            "Orzhov Enforcer",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Orzhov Enforcer — afterlife 1: create a 1/1 white and black Spirit creature token with flying",
				effects.KnightOfTheLastBreathSpirits(1)),
		},
	})
}
