package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Torpid Moloch — Creature — Lizard {R}, 3/2 (Ravnica: City of Guilds, S82 OLD
// batch 574, #5600):
//
//	"Defender
//	 Sacrifice three lands: This creature loses defender until end of turn."
//
// Defender is the printed keyword. The cost sacrifices three lands you
// control (any lands, the Moloch's controller choosing); the effect removes
// defender from the Moloch until end of turn through the layered view, so
// it may attack that turn (CR 702.3).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "48f78611-13a7-4b5b-a52a-317ecf658ea3",
		Name:            "Torpid Moloch",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice three lands: This creature loses defender until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    effects.SacrificeN(3, "three lands", effects.Land()),
			Effect:  effects.ThisLosesKeywordUntilEndOfTurn("defender", "Torpid Moloch — loses defender until end of turn"),
		}},
	})
}
