package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wrecking Gecko — Artifact Creature — Lizard Construct {4}{G}, 5/5:
//
//	"Ward {2}
//	 {6}{G}{G}: This creature gets +4/+4 and gains trample until end of
//	 turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a9e0294-c967-4e90-b776-6d3195034a23",
		Name:         "Wrecking Gecko",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{2}"), "Wrecking Gecko — ward {2}"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{6}{G}{G}: This creature gets +4/+4 and gains trample until end of turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{6}{G}{G}"),
			Effect:  effects.ThisCreatureUntilEOT("Wrecking Gecko — +4/+4 and trample", 4, 4, "trample"),
		}},
	})
}
