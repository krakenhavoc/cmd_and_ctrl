package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Angelfire Crusader — Creature — Human Soldier Knight {3}{W}, 2/3 (Apocalypse,
// S82 OLD, #5614):
//
//	"{R}: This creature gets +1/+0 until end of turn."
//
// Firebreathing through the shared self-pump row: no {T}, so it can be
// activated any number of times.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a587b8d-fcb4-47ca-9689-610897117166",
		Name:         "Angelfire Crusader",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.SelfPumpRow("Angelfire Crusader", "{R}", 1, 0)},
	})
}
