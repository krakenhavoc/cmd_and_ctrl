package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kathari Screecher — {2}{U} Creature — Bird Soldier 2/2 (Shards of Alara, S82 OLD b516):
//
//	"Flying
//	 Unearth {2}{U} (...)"
//
// Unearth is an activated ability that works from the graveyard at sorcery
// speed (CR 702.84a): the creature returns with haste and is exiled at the next
// end step, or if it would leave the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb51bafe-5a70-4910-a42e-ef75d9dd2f0a",
		Name:            "Kathari Screecher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []effects.ActivatedAbility{effects.Unearth("{2}{U}")},
	})
}
