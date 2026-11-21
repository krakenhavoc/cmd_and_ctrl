package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wind Strider — Creature — Merfolk Wizard {4}{U}, 3/3:
//
//	"Flash
//	 Flying"
//
// Both keywords ride PrintedKeywords: flash is read by the cast-timing
// check and flying by the combat engine's block rules.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9d57cda6-9042-44bd-9785-38de9487ff85",
		Name:            "Wind Strider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
