package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brigone, Soldier of Meletis — Legendary Creature — Human Soldier
// {1}{W}, 2/2 (Foundations Jumpstart, #5464):
//
//	"Vigilance
//	 Heroic — Whenever you cast a spell that targets Brigone, put a
//	 +1/+1 counter on Brigone.
//	 {T}, Remove a +1/+1 counter from Brigone: Draw a card."
//
// Bloodcrazed Hoplite's heroic counter, and Benevolent Hydra's cost: the
// counter is removed from Brigone as the ability is activated
// (RemoveCountersFromThis), so with none on it the ability can't be
// activated, and a summoning-sick Brigone can't tap for it (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f328dfb1-0737-478b-bd40-dfcb9a35e5a5",
		Name:            "Brigone, Soldier of Meletis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			heroic("Brigone, Soldier of Meletis — a +1/+1 counter on Brigone (heroic)", thisGetsAPlusOneCounter),
		},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Remove a +1/+1 counter from Brigone: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
			Cost:    Plus(TapCost(), RemoveCountersFromThis(game.CounterPlusOne, 1)),
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
