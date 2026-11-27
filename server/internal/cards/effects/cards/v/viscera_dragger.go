package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Viscera Dragger — {3}{B} Creature — Zombie Ogre Warrior 3/3 (Shards of Alara, S82 OLD b516):
//
//	"Cycling {2}
//	 Unearth {1}{B}"
//
// Unearth is an activated ability that works from the graveyard at sorcery
// speed (CR 702.84a): the creature returns with haste and is exiled at the next
// end step, or if it would leave the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "77630700-3828-4063-95a7-f6283f0f9da5",
		Name:         "Viscera Dragger",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}"), effects.Unearth("{1}{B}")},
	})
}
