package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wyrm's Crossing Patrol — Creature — Human Soldier {W}, 1/1 (Commander
// Legends: Battle for Baldur's Gate, #5948):
//
//	"Myriad (Whenever this creature attacks, for each opponent other than
//	 defending player, you may create a token copy that's tapped and
//	 attacking that player or a planeswalker they control. Exile the tokens
//	 at end of combat.)"
//
// Myriad is the shared trigger (myriad.go, CR 702.116), as on Banshee of the
// Dread Choir.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fad2d14f-8825-4b08-8804-010d9659f6dc",
		Name:         "Wyrm's Crossing Patrol",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.GrantedMyriad("Wyrm's Crossing Patrol"),
		},
	})
}
