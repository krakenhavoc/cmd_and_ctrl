package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Relentless Raptor — Creature — Dinosaur {R}{W}, 3/3 (Rivals of Ixalan,
// #5531):
//
//	"Vigilance
//	 This creature attacks or blocks each combat if able."
//
// Iron Golem's text: vigilance is printed data, and the attack and block
// requirements are the engine's, enforced as attackers and blockers are
// declared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "4727db28-98e9-453d-90c3-2d51365d32fc",
		Name:            "Relentless Raptor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Static: []game.StaticAbility{
			effects.AttacksEachCombat(),
			effects.BlocksEachCombat(),
		},
	})
}
