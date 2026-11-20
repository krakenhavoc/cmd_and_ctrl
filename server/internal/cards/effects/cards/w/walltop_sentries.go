package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walltop Sentries — Creature — Human Soldier Ally {2}{G}, 2/3 (Avatar:
// The Last Airbender, #4459):
//
//	"Reach, deathtouch
//	 When this creature dies, if there's a Lesson card in your graveyard,
//	 you gain 2 life."
//
// An intervening "if" (CR 603.4): checked as the creature dies, so with no
// Lesson card in the graveyard the ability never goes on the stack, and
// again as it resolves, so a Lesson card exiled in response stops the
// gain (the ruling). Both halves are avB16WhenThisDiesIfLesson.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fb3a0910-a582-41ef-b5b9-3cda1f1cd5ce",
		Name:            "Walltop Sentries",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			effects.AvB16WhenThisDiesIfLesson("Walltop Sentries — you gain 2 life", effects.GainLife{Amount: 2}),
		},
	})
}
