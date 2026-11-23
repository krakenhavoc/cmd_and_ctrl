package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Agent of Kotis — Creature — Human Rogue {1}{U}, 2/1 (Tarkir:
// Dragonstorm, #4463):
//
//	"Renew — {3}{U}, Exile this card from your graveyard: Put two +1/+1
//	 counters on target creature. Activate only as a sorcery."
//
// A vanilla 2/1 on the battlefield; renew is renew.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "584fe1e4-029d-43fe-8c9e-6995dd392d65",
		Name:         "Agent of Kotis",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			effects.Renew("{3}{U}", "Put two +1/+1 counters on target creature.",
				effects.TargetCreature("target creature"), effects.RenewCountersOnTarget(game.CounterPlusOne, 2)),
		},
	})
}
