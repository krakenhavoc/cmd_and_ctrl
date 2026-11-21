package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Moonveil Dragon — Creature — Dragon {3}{R}{R}{R}, 5/5 (Dark Ascension, #5571):
//
//	"Flying
//	 {R}: Each creature you control gets +1/+0 until end of turn."
//
// Dawnwing Marshal's team pump at +1/+0: the creatures you control as the
// ability resolves get it (CR 611.2c), the Dragon included. It declares
// no TeamPump: TestTeamPumpRowsDeclareTheirPump reads only "Creatures you
// control get" texts.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e0cd4619-9a83-41f2-a349-358feaf2ef2f",
		Name:            "Moonveil Dragon",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Each creature you control gets +1/+0 until end of turn.",
			Cost:    effects.ManaCost("{R}"),
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Effect: effects.Do(effects.BoostUntilEOT{
				Match: effects.And(effects.Creature(), effects.YouControl()), Power: 1,
				Label: "Moonveil Dragon — each creature you control gets +1/+0",
			}),
		}},
	})
}
