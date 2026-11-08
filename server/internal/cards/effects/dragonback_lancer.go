package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dragonback Lancer — Creature — Human Soldier {3}{W}, 3/3 (Tarkir:
// Dragonstorm, #4463):
//
//	"Flying
//	 Mobilize 1 (Whenever this creature attacks, create a tapped and
//	 attacking 1/1 red Warrior creature token. Sacrifice it at the
//	 beginning of the next end step.)"
//
// Mobilize is mobilize.go (CR 702.181a), Shock Brigade's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7875c063-19c3-4e6e-81d8-1a378a4e16e5",
		Name:            "Dragonback Lancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered:       []game.TriggeredAbility{Mobilize("Dragonback Lancer", 1)},
	})
}
