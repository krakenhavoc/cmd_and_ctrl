package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elbrus, the Binding Blade // Withengar Unbound — Legendary Artifact —
// Equipment {7} // Legendary Creature — Demon 13/13:
//
//	Elbrus: "Equipped creature gets +1/+0.
//	 When equipped creature deals combat damage to a player, unattach Elbrus,
//	 then transform it.
//	 Equip {1}"
//	Withengar Unbound: "Flying, intimidate, trample
//	 Whenever a player loses the game, put thirteen +1/+1 counters on
//	 Withengar Unbound."
//
// The trigger is mandatory. It unattaches first and only then transforms, as
// printed; the transform is ignored if the Equipment has turned over since the
// trigger went on the stack (CR 701.28f, TransformThis). Withengar's counters ride
// the same elimination event Blood Tyrant reads, once per player who loses.
func init() {
	const oracle = "da5881ca-b86e-4582-ae0e-1570c59773ae"
	effects.Register(effects.Spec{
		OracleID:     oracle,
		Name:         "Elbrus, the Binding Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttachedCreatureDealtCombatDamageToPlayer(ev, source, g)
			}, "Elbrus, the Binding Blade — unattach it, then transform it", func(g *game.Game, item *game.StackItem) error {
				if err := effects.UnattachSource(g, item); err != nil {
					return err
				}
				return effects.TransformThis{}.Apply(effects.NewContext(g, item))
			}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{1}")},
	})
	effects.Register(effects.Spec{
		OracleID:        oracle + "#1",
		Name:            "Withengar Unbound",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "intimidate", "trample"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventPlayerEliminated, func(_ game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return true
			}, "Withengar Unbound — put thirteen +1/+1 counters on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 13})),
		},
	})
}
