package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Acolyte of Aclazotz — Creature — Vampire Cleric {2}{B}, 1/4 (The Lost
// Caverns of Ixalan, #4470):
//
//	"{T}, Sacrifice another creature or artifact: Each opponent loses 1
//	 life and you gain 1 life."
//
// Bartolomé del Presidio's cost with a tap added: both are paid as the
// ability is activated (CR 602.2b), so the sacrificed permanent is gone
// before anyone can respond. "Another" is object identity, so the
// Acolyte can't pay with itself. The tap makes it subject to summoning
// sickness (CR 302.6). The drain is the footprinted pair Dark Pool uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "324d2aa8-6dea-4b0d-898d-1e52ed187565",
		Name:         "Acolyte of Aclazotz",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature or artifact: Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost: game.AbilityCost{
				Tap: true,
				SacrificeOther: Another(sacrificeSpec("another creature or artifact",
					Or(Creature(), Artifact()))),
			},
			Effect: Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1}),
		}},
	})
}
