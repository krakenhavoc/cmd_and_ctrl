package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "fb3a0910-a582-41ef-b5b9-3cda1f1cd5ce",
		Name:            "Walltop Sentries",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "deathtouch"},
		Triggered: []game.TriggeredAbility{
			avB16WhenThisDiesIfLesson("Walltop Sentries — you gain 2 life", GainLife{Amount: 2}),
		},
	})
}
