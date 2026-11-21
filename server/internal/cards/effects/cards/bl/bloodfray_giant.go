package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodfray Giant — Creature — Giant {2}{R}{R}, 4/3 (Return to Ravnica, #5566):
//
//	"Trample / Unleash (You may have this creature enter with a +1/+1
//	 counter on it. It can't block as long as it has a +1/+1 counter on
//	 it.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Unleash is the engine's keyword (ADR 0109 §10): the entering "may" and
// the can't-block while it has a +1/+1 counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5f83fd2f-c4cb-4bcf-a4db-0238df251740",
		Name:            "Bloodfray Giant",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample", game.KeywordUnleash},
	})
}
