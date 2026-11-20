package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abzan Kin-Guard — Creature — Human Warrior {3}{G}, 3/3 (Fate Reforged,
// #5552):
//
//	"This creature has lifelink as long as you control a white or black
//	 permanent."
//
// A layer 6 grant (CR 613.1f) whose condition is read on every recompute
// (frfSelfWhileYouControlEither), so the lifelink comes and goes with your
// white and black permanents.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b292d183-1ba9-40b4-9cba-bd70bf1380ef",
		Name:         "Abzan Kin-Guard",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{abzanKinGuardLifelink()},
	})
}

// abzanKinGuardLifelink is the conditional lifelink, labelled as printed.
func abzanKinGuardLifelink() game.StaticAbility {
	s := KeywordGrant(frfSelfWhileYouControlEither("W", "B"), "lifelink")
	s.Label = "This creature has lifelink as long as you control a white or black permanent."
	return s
}
