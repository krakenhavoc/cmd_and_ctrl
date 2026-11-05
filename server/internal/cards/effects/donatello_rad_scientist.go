package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Donatello, Rad Scientist — Legendary Creature — Mutant Ninja Turtle
// {5}{U}, 5/6 (TMNT Commander, #5011):
//
//	"Vigilance
//	 When Donatello enters, tap up to three target creatures your
//	 opponents control. Put a stun counter on each of them."
//
// Vigilance rides PrintedKeywords. The enter trigger takes zero to three
// creatures your opponents control; each still legal as it resolves is
// tapped and gets a stun counter (CR 122.1d) — Hapatra's frost effect,
// shared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "671209d5-8b56-4ce7-94a0-d770d698c0df",
		Name:            "Donatello, Rad Scientist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Donatello, Rad Scientist — tap up to three target creatures your opponents control and stun them",
					rfCreatureCHapatraFrostEffect),
				TargetCreature("up to three target creatures your opponents control", OpponentControls()).WithCount(0, 3)),
		},
	})
}
