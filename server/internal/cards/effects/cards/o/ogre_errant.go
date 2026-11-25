package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Errant — Creature — Ogre Knight {3}{R}, 3/4 (Throne of Eldraine, S82
// old sets):
//
//	"Whenever this creature attacks, another target attacking Knight gains
//	 menace until end of turn."
//
// The target is chosen as the trigger goes on the stack ("another" excludes the
// Ogre itself, who is a Knight too), and a target that is gone gains nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fbaa6819-3372-4f0b-abb8-773805135153",
		Name:         "Ogre Errant",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OldB147AttackerGainsUntilEndOfTurn("Ogre Errant — another target attacking Knight gains menace until end of turn",
				"another target attacking Knight", "menace", effects.Subtype("Knight")),
		},
	})
}
