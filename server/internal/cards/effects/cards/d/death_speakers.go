package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Death Speakers — Creature — Human Cleric {W}, 1/1 (Homelands, #5635):
//
//	"Protection from black"
//
// Protection from black is an engine keyword carried as printed data:
// black spells and abilities from black sources can't target it, and
// damage from black sources is prevented (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4a1db129-cc7c-410a-a41e-d633e30cfdad",
		Name:            "Death Speakers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from black"},
	})
}
