package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "324d2aa8-6dea-4b0d-898d-1e52ed187565",
		Name:         "Acolyte of Aclazotz",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}, Sacrifice another creature or artifact: Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost: game.AbilityCost{
				Tap: true,
				SacrificeOther: effects.Another(effects.SacrificeSpec("another creature or artifact",
					effects.Or(effects.Creature(), effects.Artifact()))),
			},
			Effect: effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
