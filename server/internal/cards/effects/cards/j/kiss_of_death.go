package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kiss of Death — Sorcery {4}{B}{B} (Portal Second Age, S82 older sets batch 736, #5625):
//
//	"Kiss of Death deals 4 damage to target opponent or planeswalker. You gain 4 life."
//
// The damage, then the life (CR 608.2c); nothing at all when the only target is illegal on resolution (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c2c25406-e84a-47b0-8a46-fe60f9b0cffa",
		Name:         "Kiss of Death",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{LifeGain: 4, Targets: effects.ForTargets(effects.DamageToTarget(0, 4)).Targets},
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		OnResolve:    effects.S82OldB736DamageThenGain(4),
	})
}
