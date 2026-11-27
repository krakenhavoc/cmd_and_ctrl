package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Sprite-Chaser — Creature — Goblin Warrior {1}{R}, 1/2 (Lorwyn,
// S82 OLD b545, #5593):
//
//	"As long as you control a Faerie, this creature gets +1/+1 and has
//	 flying."
//
// Kithkin Greatheart's shape with a Faerie and flying: +1/+1 in layer 7c,
// the keyword in layer 6, one condition read on every recompute. A
// changeling you control is a Faerie (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e52d8e5c-59cf-48bb-80f7-f7996c8a81bc",
		Name:         "Boggart Sprite-Chaser",
		Completeness: effects.CompletenessFull,
		Static:       effects.S82OldB545SelfWhileYouControlSubtype("Faerie", 1, 1, "flying", "As long as you control a Faerie, this creature gets +1/+1 and has flying."),
	})
}
