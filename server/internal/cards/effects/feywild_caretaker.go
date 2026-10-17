package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Feywild Caretaker — Creature — Orc Wizard {4}{U}, 3/4:
//
//	"When this creature enters, you take the initiative.
//	 At the beginning of your end step, if you have the initiative,
//	 create a 1/1 blue Faerie Dragon creature token with flying."
//
// The ETB is TakeTheInitiative (ADR 0203, #4096); the end-step clause
// is an intervening "if" (CR 603.4), IfYouHaveTheInitiative.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f319b671-0388-48e8-ba98-04161a3e7f1a",
		Name:         "Feywild Caretaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouTakeTheInitiative("Feywild Caretaker"),
			IfYouHaveTheInitiative(AtYourEndStep("Feywild Caretaker — create a 1/1 blue Faerie Dragon with flying",
				Do(CreateToken{Template: TokenCard("1/1 blue Faerie Dragon with flying"), N: 1}))),
		},
	})
}
