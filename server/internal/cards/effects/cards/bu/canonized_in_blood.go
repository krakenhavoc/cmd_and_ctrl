package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Canonized in Blood — Enchantment {1}{B} (The Lost Caverns of Ixalan,
// #4470):
//
//	"At the beginning of your end step, if you descended this turn, put a
//	 +1/+1 counter on target creature you control.
//	 {5}{B}{B}, Sacrifice this enchantment: Create a 4/3 white and black
//	 Vampire Demon creature token with flying."
//
// The end-step trigger is Deep Goblin Skulltaker's
// (AtYourEndStepIfDescended: CR 700.11, an intervening if per CR 603.4)
// with Malamet Veteran's target: the creature you control is chosen as
// the trigger goes on the stack (CR 603.3d), and with none the trigger is
// removed. The sacrifice is part of the cost (CR 602.2b), so the
// enchantment is gone before anyone can respond; the token is Vito,
// Fanatic of Aclazotz's Vampire Demon. The ability may be activated at
// instant speed, so it declares the blocker it makes.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d491995f-e9b2-4187-852c-8a4c2495eaca",
		Name:         "Canonized in Blood",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.AtYourEndStepIfDescended("Canonized in Blood — put a +1/+1 counter on target creature you control",
				effects.PlusOneCounterOnChosenTarget),
				effects.TargetCreature("target creature you control", effects.YouControl())),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}{B}{B}, Sacrifice this enchantment: Create a 4/3 white and black Vampire Demon creature token with flying.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker, Tokens: 1},
			Cost:    effects.Plus(effects.ManaCost("{5}{B}{B}"), effects.SacrificeThis()),
			Effect:  effects.Do(effects.CreateToken{Template: effects.TokenCard("4/3 white and black Vampire Demon with flying"), N: 1}),
		}},
	})
}
