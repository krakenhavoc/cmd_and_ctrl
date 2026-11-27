package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravenous Skirge — Creature — Phyrexian Imp {2}{B} 1/1 (Urza's Saga, #5624):
//
//	"Flying
//	 Whenever this creature attacks, it gets +2/+0 until end of turn."
//
// No simplification.
func init() {
	const label = "Ravenous Skirge — it gets +2/+0 until end of turn"
	effects.Register(effects.Spec{
		OracleID:        "b8ae71b3-8ee7-49a0-87ea-7e938166ccb3",
		Name:            "Ravenous Skirge",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttack, effects.ThisAttacked, label, effects.ThisGetsUntilEndOfTurn(2, 0, label)),
		},
	})
}
