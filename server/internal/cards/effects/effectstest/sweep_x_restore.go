package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_x_restore_test.go — #4253, the final seams (part X): each moved
// card paused on a prompt, restored STRICTLY (the boot path) and answered on
// the restored table:
//
//   - the damage continuations as data: a keyed Each walk
//     (Game.DealDamageEachKeyedForEffect, the walk riding each leg's damage
//     tail) under Ranger's Firebrand's tempt and Creeping Bloodsucker's
//     shared "gain life equal to the damage dealt", and the keyed creature
//     door (Game.DealDamageToCreatureKeyedForEffect) under Ballista
//     Wielder's "can't block";
//   - the keyed proliferate told what got a counter
//     (Game.ProliferateChoosingThenKeyedForEffect): Ripples of Potential;
//   - PutFromLibraryOntoBattlefield's pick with its set rule registered
//     (RegisterPutRule) and its counters as data (PutCounters): Ojer Kaslem
//     and Turntimber Symbiosis.

// damageOrderProbeOracles are two test cards, each "if a source would deal
// damage to you or a permanent you control, it deals that much instead":
// two declared effects, so such damage asks its controller for their order
// (CR 616.1). Catalog rows, so a restored table gathers them again.
var damageOrderProbeOracles = []string{"test-4253-x-damage-order-a", "test-4253-x-damage-order-b"}

func init() {
	for i, oracle := range damageOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Damage Order Probe X" + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, self *game.Card) bool {
					if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 || ev.IsCombatDamage || self == nil {
						return false
					}
					if ev.DamageTarget == self.Controller {
						return true
					}
					c, ok := g.LookupCardForEffect(ev.DamageTarget)
					return ok && c.Controller == self.Controller && c.InstanceID != self.InstanceID
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Damage Order Probe X" + string(rune('A'+i)),
			}},
		})
	}
}

// seedDamageOrderProbes gives `who` the two probes.
func SeedDamageOrderProbes(g *game.Game, who uuid.UUID) {
	SeedOrderProbes(g, who, damageOrderProbeOracles, "Damage Order Probe X")
}
