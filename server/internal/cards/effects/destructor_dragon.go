package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Destructor Dragon — Creature — Dragon {4}{G}{G}, 4/4 (Fate Reforged,
// #5552):
//
//	"Flying
//	 When this creature dies, destroy target noncreature permanent."
//
// Flying is printed. Angel of Despair's destroy on a dies trigger, with the
// clause narrowed to a noncreature permanent; the target is chosen as the
// trigger goes on the stack and re-checked as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b0104d66-5c28-482f-838a-416aabaeaf43",
		Name:            "Destructor Dragon",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Destructor Dragon — destroy target noncreature permanent", destroyChosenPermanent),
				TargetPermanent("target noncreature permanent", Noncreature())),
		},
	})
}
