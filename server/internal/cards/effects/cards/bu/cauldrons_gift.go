package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cauldron's Gift — Sorcery {4}{B} (Throne of Eldraine, #5516):
//
//	"Adamant — If at least three black mana was spent to cast this spell,
//	 mill four cards.
//	 You may choose a creature card in your graveyard. If you do, return it to
//	 the battlefield with an additional +1/+1 counter on it."
//
// The adamant mill first, then the choice, asked as the spell resolves
// (CR 608.2c) over the creature cards then in the graveyard, so a card the
// mill put there is a legal pick. It is not a target. The chosen card enters
// with the counter as a CR 614.1c clause carried on its entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "152bbb62-195b-4de1-90c6-10dd953f8c4b",
		Name:           "Cauldron's Gift",
		Completeness:   effects.CompletenessFull,
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"B"},
		Purpose:        game.Purpose{Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.EldB1IfAdamantSpent("B", effects.MillCards{N: 4}),
			effects.Instruction(effects.OldB140CauldronsGiftChoose),
		),
	})
}
