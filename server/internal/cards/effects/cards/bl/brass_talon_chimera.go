package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Brass-Talon Chimera — Artifact Creature — Chimera {4}, 2/2 (Visions, #5631):
//
//	"first strike
//	 Sacrifice this creature: Put a +2/+2 counter on target Chimera
//	 creature. It gains first strike. (This effect lasts indefinitely.)"
//
// first strike is an engine keyword carried as printed data. The ability is the
// Visions Chimeras' shared one (ChimeraSacrificeAbility): the counter goes
// on the target, and the keyword is granted with no stated duration
// (CR 611.2a), pinned to that object. The Chimera itself is sacrificed as
// the cost, so it is never its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2e114d15-4356-4cef-8f89-260ab4c8d4b9",
		Name:            "Brass-Talon Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []effects.ActivatedAbility{effects.ChimeraSacrificeAbility("first strike")},
	})
}
