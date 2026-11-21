package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Denethor, Ruling Steward — Legendary Creature — Human Noble {1}{W}{B}, 2/4
// (The Lord of the Rings: Tales of Middle-earth):
//
//	"At the beginning of your end step, if a creature died under your
//	 control this turn, create a 1/1 white Human Soldier creature token.
//	 {2}, Sacrifice another creature: Each opponent loses 1 life and you
//	 gain 1 life."
//
// The end-step trigger is the shared intervening-if (AtYourEndStepIfACreatureDied,
// CR 603.4): checked as the step begins and again as the ability resolves
// against the turn's tally, which counts tokens (CR 700.4) and triggers once
// however many died. The activated ability is Acolyte of Aclazotz's drain
// behind a mana and a sacrifice cost ("another" as an object).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1b82312c-f748-4c36-8934-ecbcfc10cc7d",
		Name:         "Denethor, Ruling Steward",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtYourEndStepIfACreatureDied("Denethor, Ruling Steward — create a 1/1 white Human Soldier creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Human Soldier"), N: 1})),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, Sacrifice another creature: Each opponent loses 1 life and you gain 1 life.",
			Purpose: game.Purpose{Answers: game.AnswerSacOutlet, LifeGain: 1},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.SacrificeAnotherN(1, "another creature", effects.Creature())),
			Effect:  effects.Do(effects.EachOpponentLosesLife{N: 1}, effects.GainLife{Amount: 1}),
		}},
	})
}
