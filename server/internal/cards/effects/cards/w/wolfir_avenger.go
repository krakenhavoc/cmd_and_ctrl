package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wolfir Avenger — {1}{G}{G} Creature — Wolf Warrior 3/3 (Avacyn Restored, #5570):
//
//	"Flash (You may cast this spell any time you could cast an
//	 instant.)
//	 {1}{G}: Regenerate this creature."
//
// Flash and the shared regeneration row (regenerate_this.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "83deda94-66e8-4655-9cd5-6d97d809946d",
		Name:            "Wolfir Avenger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated:       []effects.ActivatedAbility{effects.RegenerateThisRow("{1}{G}")},
	})
}
