package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "584fe1e4-029d-43fe-8c9e-6995dd392d65",
		Name:         "Agent of Kotis",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			Renew("{3}{U}", "Put two +1/+1 counters on target creature.",
				TargetCreature("target creature"), renewCountersOnTarget(game.CounterPlusOne, 2)),
		},
	})
}
