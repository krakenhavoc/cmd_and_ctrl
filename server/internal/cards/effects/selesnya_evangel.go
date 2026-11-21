package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4a2fea39-99d6-4766-9f2c-0a63925349cb",
		Name:         "Selesnya Evangel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}, Tap an untapped creature you control: Create a 1/1 green Saproling creature token.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost: Plus(ManaCost("{1}"), TapCost(),
				TapAnotherUntapped("an untapped creature you control", Creature(), YouControl())),
			Effect: Do(CreateToken{Template: TokenCard("1/1 green Saproling"), N: 1}),
		}},
	})
}
