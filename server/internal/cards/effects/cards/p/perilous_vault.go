package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "7acd7ef5-0e26-436d-962a-1a391f81c2e8",
		Name:         "Perilous Vault",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, {T}, Exile this artifact: Exile all nonland permanents.",
			Purpose: game.Purpose{Answers: game.AnswerRemove, Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepExile}},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.TapCost(), effects.ExileThis()),
			Effect:  effects.Do(effects.ExileAllMatching{Match: effects.Not(effects.Land())}),
		}},
	})
}
