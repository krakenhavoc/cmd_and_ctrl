package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Benalish Trapper (Invasion, S82 OLD batch 647, #5616):
//
//	Creature — Human Soldier {1}{W}, 1/2:
//	"{W}, {T}: Tap target creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d45c6c3f-6079-40c1-9083-06f2f2431bcc",
		Name:         "Benalish Trapper",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.S82OldB647TapTargetRow("{W}", "{W}, {T}")},
	})
}
