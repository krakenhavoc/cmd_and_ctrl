package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stench of Evil — Sorcery {2}{B}{B} (Ice Age, S82 OLD, #5636):
//
//	"Destroy all Plains. For each land destroyed this way, Stench of Evil deals 1 damage to
//	 that land's controller unless they pay {2}."
//
// Flashfires's sweep with a per-land tail: each land that was actually destroyed (an
// indestructible Plains is not) asks its controller, as the land last existed, whether to pay {2}
// (CR 118.12); a "no" is 1 damage from the sorcery.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "14062abe-14df-4dbe-ad7b-530e5f6c4988",
		Name:         "Stench of Evil",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepDestroy, Partial: true}},
		OnResolve:    effects.S82OldB842StenchOfEvil,
	})
}
