package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Splatter Thug — Creature — Human Warrior {2}{R}, 2/2 (Return to Ravnica, #5566):
//
//	"First strike / Unleash (You may have this creature enter with a
//	 +1/+1 counter on it. It can't block as long as it has a +1/+1
//	 counter on it.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Unleash is the engine's keyword (ADR 0109 §10): the entering "may" and
// the can't-block while it has a +1/+1 counter.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "3baee36b-d64f-448c-ae45-c1ec124c92ec",
		Name:            "Splatter Thug",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", game.KeywordUnleash},
	})
}
