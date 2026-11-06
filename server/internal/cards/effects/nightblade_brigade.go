package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nightblade Brigade — Creature — Goblin Soldier {2}{B}, 1/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Deathtouch
//	 Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)
//	 When this creature enters, surveil 1. (Look at the top card of your
//	 library. You may put it into your graveyard.)"
//
// Mobilize is mobilize.go (CR 702.181a), Shock Brigade's. The entry
// trigger is the ordinary surveil (CR 701.25a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "36d04190-7ce3-431a-b18e-3a1895ec6fc7",
		Name:            "Nightblade Brigade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{
			Mobilize("Nightblade Brigade", 1),
			WhenThisEnters("Nightblade Brigade — surveil 1", Do(Surveil{N: 1})),
		},
	})
}
