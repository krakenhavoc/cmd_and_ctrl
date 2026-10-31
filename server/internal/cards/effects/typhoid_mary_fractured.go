package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Typhoid Mary, Fractured — Legendary Creature — Mutant Villain {1}{B}{R},
// 3/3 (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Typhoid Mary attacks, choose one at random. If you discarded a
//	 card this turn, you choose one instead.
//	 • Mary — Create a Treasure token.
//	 • Typhoid Mary — Draw a card.
//	 • Bloody Mary — Each opponent loses 2 life and you gain 2 life."
//
// ChooseOneAtRandom (CR 700.2, #3399) draws the bullet as the trigger goes on
// the stack, unless the controller has discarded a card this turn
// (YouDiscardedACardThisTurn, the turn tally's CardsDiscarded, the chooser's
// own count), in which case ChosenByYouInsteadIf hands them the ordinary
// "choose one" prompt. The condition is read as the ability is put on the
// stack (CR 603.3c), when modes are chosen.
//
// No simplification.
func init() {
	const name = "Typhoid Mary, Fractured"
	t := WheneverThisAttacks(name+" — choose one at random: a Treasure, a card, or a drain of 2",
		func(*game.Game, *game.StackItem) error { return nil })
	t.Modes = ChooseOneAtRandom(
		ModeSteps("Mary — Create a Treasure token.", nil,
			CreateToken{Template: TreasureToken(), N: 1}),
		ModeSteps("Typhoid Mary — Draw a card.", nil,
			DrawCards{N: 1}),
		ModeSteps("Bloody Mary — Each opponent loses 2 life and you gain 2 life.", nil,
			EachOpponentLosesLife{N: 2}, GainLife{Amount: 2}),
	).ChosenByYouInsteadIf(YouDiscardedACardThisTurn)
	Register(Spec{
		OracleID:     "867b3923-692c-4444-a071-518a290e43d8",
		Name:         name,
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
