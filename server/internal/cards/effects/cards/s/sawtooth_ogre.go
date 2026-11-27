package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Sawtooth Ogre — {2}{R}{R} Creature — Ogre, 3/3 (Weatherlight, S82 OLD batch 771,
// #5629):
//
//	"Whenever this creature blocks or becomes blocked by a creature, this creature
//	 deals 1 damage to that creature at end of combat."
//
// Corrosive Ooze's per-pair block trigger (CR 509.3b, 509.3d) with a CR 603.7
// delayed trigger at the beginning of the end of combat step naming that
// creature (S82OldB771SawtoothTrigger); the Ogre deals the damage if the creature
// is still on the battlefield then, and as it last existed if the Ogre is not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "473c9f27-2091-4d87-b279-64b3662c9af8",
		Name:         "Sawtooth Ogre",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.S82OldB771SawtoothTrigger()},
	})
}
