package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dovin's Veto — Instant {W}{U} (EDHREC rank 239):
//
//	"This spell can't be countered.
//	 Counter target noncreature spell."
//
// Negate with the S23 "can't be countered" rider — the same two
// shapes Counterflux and Dragonlord Dromoka already compose.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1b388371-f9ef-45b4-82a3-ca20a8cd7807",
		Name:            "Dovin's Veto",
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetSpell("target noncreature spell", effects.Noncreature()),
		OnResolve:       effects.CounterTheTargetSpell,
	})
}
