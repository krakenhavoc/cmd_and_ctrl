package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Staunch Throneguard — Artifact Creature — Construct {5}, 2/5 (Commander
// Legends, #5507):
//
//	"Vigilance
//	 When this creature enters, you become the monarch."
//
// Vigilance is the printed keyword; the enters trigger is every monarch
// card's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c8131fb3-3146-4434-9e4b-ac88ab3cbe37",
		Name:            "Staunch Throneguard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouBecomeTheMonarch("Staunch Throneguard"),
		},
	})
}
