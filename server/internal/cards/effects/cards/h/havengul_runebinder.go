package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Havengul Runebinder — Creature — Human Wizard {2}{U}{U}, 2/2:
//
//	"{2}{U}, {T}, Exile a creature card from your graveyard: Create a
//	 2/2 black Zombie creature token, then put a +1/+1 counter on each
//	 Zombie creature you control."
//
// The exile is a cost (CR 118.3), so an empty graveyard refuses the
// activation. The counters go on every Zombie its controller controls once
// the new token is there, the token included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e91e7f62-0788-445e-ba93-2c9d233ec3ab",
		Name:         "Havengul Runebinder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}, {T}, Exile a creature card from your graveyard: Create a 2/2 black Zombie creature token, then put a +1/+1 counter on each Zombie creature you control.",
			Purpose: game.Purpose{Tokens: 1, CreatureTokens: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost(), effects.ExileFromGraveyard(1, "a creature card", effects.IsCreatureCard)),
			Effect: effects.Do(
				effects.CreateToken{Template: effects.BlackZombieToken(), N: 1},
				effects.Instruction(effects.S82RprB31CounterOnEachZombieYouControl),
			),
		}},
	})
}
