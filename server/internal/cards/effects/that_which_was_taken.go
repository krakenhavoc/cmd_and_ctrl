package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// That Which Was Taken — Legendary Artifact {5} (Betrayers of Kamigawa,
// #5602):
//
//	"{4}, {T}: Put a divinity counter on target permanent other than That
//	 Which Was Taken.
//	 Each permanent with a divinity counter on it has indestructible."
//
// Kindred Boon's two halves without its type or controller limits: the
// activation targets any other permanent (Another), and the layer 6 grant
// covers every permanent with a divinity counter, whoever controls it and
// whichever card put the counter there, for as long as the artifact is on
// the battlefield.
//
// No simplification.
func init() {
	indestructible := KeywordGrant(func(target *game.Card, _ *game.Game, _ *game.Card) bool {
		return target.Counters["divinity"] > 0
	}, "indestructible")
	indestructible.Label = "Each permanent with a divinity counter on it has indestructible."
	Register(Spec{
		OracleID:     "b3c5b7b6-8002-41f2-87fc-1ed7f9e103ee",
		Name:         "That Which Was Taken",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{indestructible},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Put a divinity counter on target permanent other than That Which Was Taken.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Targets: Another(TargetPermanent("target permanent other than That Which Was Taken")),
			Effect:  putACounterOnTheTarget("divinity"),
		}},
	})
}
