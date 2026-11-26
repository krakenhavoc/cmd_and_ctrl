package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Domri's Ambush — Sorcery {R}{G} (War of the Spark):
//
//	"Put a +1/+1 counter on target creature you control. Then that creature
//	 deals damage equal to its power to target creature or planeswalker you
//	 don't control."
//
// Felling Blow's resolution with a planeswalker as a legal victim: the counter
// is its own step, so the power read is the power with the counter on it, and the
// damage is dealt only if both targets are still legal (CR 608.2b). The creature
// is the source of the damage (CR 609.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a481f910-59a8-48b0-a8e4-c24675e1a67d",
		Name:         "Domri's Ambush",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.TargetPermanent("target creature or planeswalker you don't control", effects.Or(effects.Creature(), effects.Planeswalker()), effects.OpponentControls()),
		),
		OnResolve: effects.CounterOnYourCreatureThenBite(),
	})
}
