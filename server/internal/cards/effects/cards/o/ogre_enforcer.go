package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ogre Enforcer — Creature — Ogre {3}{R}{R}, 4/4 (Visions, S82 older sets batch
// 792, #5631):
//
//	"This creature can't be destroyed by lethal damage unless lethal damage
//	 dealt by a single source is marked on it."
//
// The state-based action for lethal damage (CR 704.5g) reads the per-source
// split of the damage marked (Card.MarkedBySource, #6209) for a creature with
// this static: damage that is lethal only as a sum does not destroy it. Deathtouch
// (CR 704.5h) and "destroy" effects are not lethal damage and still do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:               "86cc83d8-4a70-4b97-aa4d-160afd4cd6cb",
		Name:                   "Ogre Enforcer",
		Completeness:           effects.CompletenessFull,
		SingleSourceLethalOnly: effects.SingleSourceLethalOnly("This creature can't be destroyed by lethal damage unless lethal damage dealt by a single source is marked on it."),
	})
}
