package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleetfeather Cockatrice — Creature — Cockatrice {3}{G}{U}, 3/3 (Journey into
// Nyx, S82 OLD, #5558):
//
//	"Flash
//	 Flying, deathtouch
//	 {5}{G}{U}: Monstrosity 3. (If this creature isn't monstrous, put three
//	 +1/+1 counters on it and it becomes monstrous.)"
//
// Printed keywords and the monstrosity constructor.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d5006683-f1c4-4a18-bf64-2d075372534c",
		Name:            "Fleetfeather Cockatrice",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying", "deathtouch"},
		Activated:       []effects.ActivatedAbility{effects.Monstrosity(effects.ManaCost("{5}{G}{U}"), 3)},
	})
}
