package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fire-Field Ogre — {1}{U}{B}{R} Creature — Ogre Mutant 4/2 (Shards of Alara, S82 OLD b516):
//
//	"First strike
//	 Unearth {U}{B}{R} (...)"
//
// Unearth is an activated ability that works from the graveyard at sorcery
// speed (CR 702.84a): the creature returns with haste and is exiled at the next
// end step, or if it would leave the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "343a258e-ea46-4acb-8543-9f869e43bc59",
		Name:            "Fire-Field Ogre",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{U}{B}{R}")},
	})
}
