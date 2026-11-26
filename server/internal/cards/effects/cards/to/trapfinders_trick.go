package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trapfinder's Trick — Sorcery {1}{U} (Zendikar, #5583):
//
//	"Target player reveals their hand and discards all Trap cards."
//
// The hand is revealed, then every Trap card in it is discarded as one
// instruction (CR 701.9).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9792ae99-7b3a-46cb-bd22-0b33c499ce1a",
		Name:         "Trapfinder's Trick",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.OldB497TrapfindersTrick,
	})
}
