package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dunland Crebain — Creature — Bird Horror {2}{B}, 1/1 (The Lord of the
// Rings: Tales of Middle-earth):
//
//	"Flying
//	 When this creature enters, amass Orcs 2."
//
// Aven Eternal's shape at two counters: the Amass primitive finds the Army
// or makes the 0/0 Orc Army token first (CR 701.47).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9dd755a5-0bef-4bde-92fa-14ef2a1a1410",
		Name:            "Dunland Crebain",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Amass: 2},
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dunland Crebain — amass Orcs 2", effects.Do(effects.Amass{Subtype: "Orc", N: 2})),
		},
	})
}
