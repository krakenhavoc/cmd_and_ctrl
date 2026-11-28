package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ivory Gargoyle — Creature — Gargoyle {4}{W}, 2/2 (Alliances, #5634):
//
//	"Flying
//	 When this creature dies, return it to the battlefield under its owner's
//	 control at the beginning of the next end step and you skip your next draw
//	 step.
//	 {4}{W}: Exile this creature."
//
// The dies trigger does two things as it resolves: it schedules the return
// (a CR 603.7 delayed trigger naming the card as the object it is in the
// graveyard, so a Gargoyle that left the graveyard is not returned, CR 400.7)
// under its OWNER's control, and it makes its controller skip their next draw
// step (CR 614.10). The exile ability is a plain activated one.
//
// No simplification.
func init() {
	const name = "Ivory Gargoyle"
	effects.Register(effects.Spec{
		OracleID:        "e8cc9f83-4e4c-4820-9a7a-f82f728e9767",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies(name+" — return it to the battlefield at the beginning of the next end step and you skip your next draw step",
				effects.OldB814ReturnItAtTheNextEndStepAndSkipYourNextDrawStep(name)),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{W}: Exile this creature.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{4}{W}"),
			Effect:  effects.OldB814ExileThisCreature,
		}},
	})
}
