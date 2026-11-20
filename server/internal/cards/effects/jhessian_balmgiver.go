package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Jhessian Balmgiver — Creature — Human Cleric {1}{W}{U}, 1/1 (Conflux,
// #5587):
//
//	"{T}: Prevent the next 1 damage that would be dealt to any target
//	 this turn.
//	 {T}: Target creature can't be blocked this turn."
//
// Benevolent Ancestor's prevention shield (CR 615.7) and Access Tunnel's
// evasion, each on a {T} row, so summoning sickness holds both back (CR
// 302.6) and only one can be used a turn.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2be283da-6d5a-442b-bd07-d06c16cb1214",
		Name:         "Jhessian Balmgiver",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Prevent the next 1 damage that would be dealt to any target this turn.",
				Purpose: game.Purpose{Answers: game.AnswerPrevent},
				Cost:    TapCost(),
				Targets: TargetAny(),
				Effect:  preventNextOneDamageToTheTarget("Jhessian Balmgiver"),
			},
			{
				Label:   "{T}: Target creature can't be blocked this turn.",
				Purpose: ForTargets(CantBeBlockedUntilEOT(0, 0, 0)),
				Cost:    TapCost(),
				Targets: TargetCreature("target creature"),
				Effect:  targetCantBeBlockedThisTurn("Jhessian Balmgiver"),
			},
		},
	})
}
