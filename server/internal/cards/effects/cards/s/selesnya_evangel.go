package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Evangel — Creature — Elf Shaman {G}{W}, 1/2 (Ravnica Remastered,
// #5948):
//
//	"{1}, {T}, Tap an untapped creature you control: Create a 1/1 green
//	 Saproling creature token."
//
// Loam Dryad's cost: the {T} taps the Evangel, so the other creature is another
// one (TapAnotherUntapped); tapping it is not the {T} symbol (CR 302.6), so a
// creature that arrived this turn pays it, and it does not target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a2fea39-99d6-4766-9f2c-0a63925349cb",
		Name:         "Selesnya Evangel",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Tap an untapped creature you control: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost: effects.Plus(effects.ManaCost("{1}"), effects.TapCost(),
				effects.TapAnotherUntapped("an untapped creature you control", effects.Creature(), effects.YouControl())),
			Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}),
		}},
	})
}
