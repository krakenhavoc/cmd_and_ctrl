package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

var consulsShieldguardIndestructiblePaid = effects.EnergyPaid("pay-unless/consuls-shieldguard-indestructible",
	effects.FirstTargetGainsUntilEndOfTurn(0, "Consul's Shieldguard — indestructible until end of turn", "indestructible"))

// Consul's Shieldguard — Creature — Dwarf Soldier {3}{W}, 3/4:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}. If you do, another
//	 target attacking creature gains indestructible until end of turn."
//
// ADR 0129 §3 (#1995): the target is chosen as the trigger goes on the
// stack (CR 603.3d); the energy is paid as it resolves (CR 118.12), and
// the grant lands on that target.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c1a3eeef-7ea5-4d58-9e6f-91bbf1c2cc8c",
		Name:         "Consul's Shieldguard",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Consul's Shieldguard", 2),
			effects.AnotherAttackerMayPayEnergy("Consul's Shieldguard", "give it indestructible",
				consulsShieldguardIndestructiblePaid),
		},
	})
}
