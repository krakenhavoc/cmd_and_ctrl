package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Super Combo — Sorcery {1}{G} (Teenage Mutant Ninja Turtles Commander,
// #5011):
//
//	"Replicate {2} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.)
//	 Target creature you control deals damage equal to its power to
//	 target creature an opponent controls."
//
// Replicate is the keyword's two halves (CR 702.56a), Changing Loyalty's:
// the optional {2}, paid any number of times (replicateMana, capped at ten
// per announcement, far beyond any Commander mana pool's need), and the
// Replicate() cast trigger, which copies the spell once per payment and
// offers each copy new targets (CR 707.10c). Each copy is countered or
// resolves on its own.
//
// The effect is Bite Down's one-sided bite (oneSidedBite) over two
// clauses: slot 0 a creature you control, slot 1 a creature an opponent
// controls. The power is read as the copy resolves, and if either target
// is illegal by then the creature deals no damage (the card's ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "68555673-2833-459a-b415-e2805f837778",
		Name:          "Super Combo",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{replicateMana("{2}", 10)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: oneSidedBite,
	})
}
