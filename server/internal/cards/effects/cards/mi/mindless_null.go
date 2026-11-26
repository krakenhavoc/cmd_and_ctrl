package cardsmi

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mindless Null — Creature — Zombie {2}{B}, 2/2 (Zendikar, #5583):
//
//	"This creature can't block unless you control a Vampire."
//
// A CR 509.1b block rule read as the creature is declared as a blocker: its
// controller must control a Vampire permanent then, whatever else it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22ff4e16-5c66-4076-94cc-32c281d42f03",
		Name:         "Mindless Null",
		Completeness: effects.CompletenessFull,
		BlockRules: []game.BlockRule{effects.CantBlockUnless(mindlessNullYouControlAVampire,
			"it can't block unless its controller controls a Vampire")},
	})
}

func mindlessNullYouControlAVampire(g *game.Game, you uuid.UUID) bool {
	return g.CountMatchingForEffect(you, []game.PermanentQuery{{Subtypes: []string{"Vampire"}}}) > 0
}
