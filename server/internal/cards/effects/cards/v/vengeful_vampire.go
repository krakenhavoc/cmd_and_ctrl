package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vengeful Vampire — Creature — Vampire {4}{B}{B}, 3/2 (Dark Ascension, #5571):
//
//	"Flying
//	 Undying (When this creature dies, if it had no +1/+1 counters on
//	 it, return it to the battlefield under its owner's control with
//	 a +1/+1 counter on it.)"
//
// Flying and undying are PrintedKeywords (#2075), as on Demonlord of
// Ashmouth.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1fcf95d7-fd70-474b-be0a-e43502633556",
		Name:            "Vengeful Vampire",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.KeywordUndying},
	})
}
