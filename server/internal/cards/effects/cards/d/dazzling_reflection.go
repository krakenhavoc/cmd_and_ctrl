package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dazzling Reflection — Instant {1}{W}:
//
//	"You gain life equal to target creature's power. The next time that creature would deal damage this turn, prevent that damage."
//
// ADR 0107 §6 (#1860). The life gain reads the creature's power as the
// spell resolves; a power of 0 or less gains nothing. The shield is
// against that creature's next instance of damage, to anything
// (CR 615.8), and names it as the object it is then (CR 400.7).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44db1ff5-7082-4da4-8f4f-4a2d2e3d7666",
		Name:         "Dazzling Reflection",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ShieldAgainstTargetCreature(game.BodyRef{}, true),
	})
}
