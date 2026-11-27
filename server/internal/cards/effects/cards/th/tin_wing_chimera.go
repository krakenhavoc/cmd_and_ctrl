package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tin-Wing Chimera — Artifact Creature — Chimera {4}, 2/2 (Visions, #5631):
//
//	"flying
//	 Sacrifice this creature: Put a +2/+2 counter on target Chimera
//	 creature. It gains flying. (This effect lasts indefinitely.)"
//
// flying is an engine keyword carried as printed data. The ability is the
// Visions Chimeras' shared one (ChimeraSacrificeAbility): the counter goes
// on the target, and the keyword is granted with no stated duration
// (CR 611.2a), pinned to that object. The Chimera itself is sacrificed as
// the cost, so it is never its own target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d5fe27f-db51-4323-aa74-b6fdb4309425",
		Name:            "Tin-Wing Chimera",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.ChimeraSacrificeAbility("flying")},
	})
}
