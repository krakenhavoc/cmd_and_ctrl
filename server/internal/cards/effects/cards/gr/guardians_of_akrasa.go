package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Guardians of Akrasa — Creature — Human Soldier {2}{W}, 0/4 (Shards of
// Alara, #5588):
//
//	"Defender
//	 Exalted"
//
// A defender can't attack (CR 702.3b), but its exalted still pumps the
// lone attacker beside it (CR 702.83a): the trigger is the controller's
// whichever of their creatures attacks alone.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7ef8df9f-3fab-44fd-b5a9-f57e54dbc38e",
		Name:            "Guardians of Akrasa",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", game.KeywordExalted},
	})
}
