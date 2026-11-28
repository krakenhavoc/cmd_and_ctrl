package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lead-Belly Chimera — Artifact Creature — Chimera {4}, 2/2 (Visions, #5631):
//
//	"trample
//	 Sacrifice this creature: Put a +2/+2 counter on target Chimera
//	 creature. It gains trample. (This effect lasts indefinitely.)"
//
// trample is an engine keyword carried as printed data. The ability is the
// Visions Chimeras' shared one (ChimeraSacrificeAbility): the counter goes
// on the target, and the keyword is granted with no stated duration
// (CR 611.2a), pinned to that object. The Chimera itself is sacrificed as
// the cost, so it is never its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e8fb2323-3eed-45a6-bfc4-44292c8b3697",
		Name:            "Lead-Belly Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.ChimeraSacrificeAbility("trample")},
	})
}
