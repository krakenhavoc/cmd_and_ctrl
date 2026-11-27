package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron-Heart Chimera — Artifact Creature — Chimera {4}, 2/2 (Visions, #5631):
//
//	"vigilance
//	 Sacrifice this creature: Put a +2/+2 counter on target Chimera
//	 creature. It gains vigilance. (This effect lasts indefinitely.)"
//
// vigilance is an engine keyword carried as printed data. The ability is the
// Visions Chimeras' shared one (ChimeraSacrificeAbility): the counter goes
// on the target, and the keyword is granted with no stated duration
// (CR 611.2a), pinned to that object. The Chimera itself is sacrificed as
// the cost, so it is never its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe7c70dd-76ff-4958-9973-6452c7f071ed",
		Name:            "Iron-Heart Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated:       []effects.ActivatedAbility{effects.ChimeraSacrificeAbility("vigilance")},
	})
}
