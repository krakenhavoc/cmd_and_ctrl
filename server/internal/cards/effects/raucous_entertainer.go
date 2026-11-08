package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raucous Entertainer — Creature — Plant Bard {1}{G}, 2/2 (Outlaws of
// Thunder Junction, #4468):
//
//	"{1}, {T}: Put a +1/+1 counter on each creature you control that
//	 entered this turn."
//
// Oakhollow Village's activation without the type filter: the set is the
// creatures you control as the ability resolves that have an entry in
// this turn's tally (game.EnteredThisTurn), read once before the first
// counter goes on (putAPlusOneCounterOnEachYouControl). A creature that
// entered on an earlier turn and is merely summoning sick does not
// count; the Entertainer itself counts on the turn it entered, if
// something has given it haste.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "534e74f6-52f4-4824-9178-cce5e4bcbcc2",
		Name:         "Raucous Entertainer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, {T}: Put a +1/+1 counter on each creature you control that entered this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPump},
			Cost:    Plus(ManaCost("{1}"), TapCost()),
			Effect:  putAPlusOneCounterOnEachYouControl(And(Creature(), rfCreatureCEnteredThisTurn())),
		}},
	})
}
