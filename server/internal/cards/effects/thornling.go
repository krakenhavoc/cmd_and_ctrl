package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thornling — Creature — Elemental Shapeshifter {3}{G}{G}, 4/4 (Conflux,
// #5587):
//
//	"{G}: This creature gains haste until end of turn.
//	 {G}: This creature gains trample until end of turn.
//	 {G}: This creature gains indestructible until end of turn.
//	 {1}: This creature gets +1/-1 until end of turn.
//	 {1}: This creature gets -1/+1 until end of turn."
//
// Five rows, each ending in the cleanup step (CR 514.2): three layer-6
// keyword grants and two layer-7c trades through the shared thisGetsRow.
// Indestructible keeps it from being destroyed, lethal damage included
// (CR 702.12b), but not from a toughness of 0 (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "55a59d01-b2cf-4c17-89c5-ce8312a85740",
		Name:         "Thornling",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			conB1ThisGainsRow("{G}", "haste", "Thornling", game.AnswerCombatGrant),
			conB1ThisGainsRow("{G}", "trample", "Thornling", game.AnswerCombatGrant),
			conB1ThisGainsRow("{G}", "indestructible", "Thornling", game.AnswerSurvive),
			thisGetsRow("{1}", 1, -1, "Thornling"),
			thisGetsRow("{1}", -1, 1, "Thornling"),
		},
	})
}
