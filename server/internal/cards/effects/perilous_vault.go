package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Perilous Vault — Artifact {4}:
//
//	"{5}, {T}, Exile this artifact: Exile all nonland permanents."
//
// Exiling the Vault is a cost, so it is gone before the ability
// resolves and is not among the permanents exiled (CR 602.2b). Every
// other nonland permanent goes, tokens included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7acd7ef5-0e26-436d-962a-1a391f81c2e8",
		Name:         "Perilous Vault",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{5}, {T}, Exile this artifact: Exile all nonland permanents.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile}},
			Cost:    Plus(ManaCost("{5}"), TapCost(), ExileThis()),
			Effect:  Do(ExileAllMatching{Match: Not(Land())}),
		}},
	})
}
