package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// sweep_k3_restore_test.go — #4253: three keyed engine forms, each moved
// card paused on a CR 616 ordering prompt, restored STRICTLY (the boot
// path) and answered on the restored table:
//
//   - "exile target spell" with its "then" as data (ExileTargetSpell.ThenTail,
//     Game.ExileSpellKeyedForEffect): Mindbreak Trap's run over its targets
//     and Aven Interrupter's plot;
//   - ExileUntil.ThenTail, riding the record of the return: Abdel Adrian's
//     count of the permanents exiled this way;
//   - damage to a player with its "then" as data, the damage event's frame
//     written to the restore point (Game.DealDamageToPlayerKeyedForEffect):
//     Screaming Nemesis's "can't gain life".

// nemesisDamageOrderProbeOracles are two test cards, each "if a source named
// Screaming Nemesis would deal damage to a player, it deals that much
// instead": two declared effects, so the Nemesis's damage to a player asks
// for their order (CR 616.1). Catalog rows, so a restored table gathers
// them again.
var NemesisDamageOrderProbeOracles = []string{"test-4253-k3-damage-order-a", "test-4253-k3-damage-order-b"}

func init() {
	for i, oracle := range NemesisDamageOrderProbeOracles {
		effects.Register(effects.Spec{
			OracleID: oracle,
			Name:     "Damage Order Probe " + string(rune('A'+i)),
			Replacements: []game.ReplacementEffect{{
				Watches: []game.EventKind{game.EventDealDamage},
				AppliesTo: func(ev *game.ReplacementEvent, g *game.Game, _ *game.Card) bool {
					if ev.Kind != game.RepEventDamage || ev.DamageAmount <= 0 {
						return false
					}
					if _, isCard := g.LookupCardForEffect(ev.DamageTarget); isCard {
						return false
					}
					c, ok := g.LookupCardForEffect(ev.DamageSource)
					return ok && c.Name == "Screaming Nemesis"
				},
				Replace: func(*game.ReplacementEvent, *game.Game, *game.Card) error { return nil },
				Label:   "Damage Order Probe " + string(rune('A'+i)),
			}},
		})
	}
}
