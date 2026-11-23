package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enormous Energy Blade — Artifact — Equipment {2}{B} (Kamigawa: Neon
// Dynasty, #5493):
//
//	"Equipped creature gets +4/+0.
//	 Whenever this Equipment becomes attached to a creature, tap that
//	 creature.
//	 Equip {2}"
//
// The pump is Skullclamp's layer 7c modify. The trigger is
// ThisBecameAttachedToACreature (#5399): moving the Blade onto a new
// creature taps it, re-attaching to the creature it already equips does
// not trigger, and the creature tapped is the host the event names,
// read as the ability resolves (a host that left does nothing, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7818e900-2f4e-4710-87a4-5a9ac170a677",
		Name:         "Enormous Energy Blade",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(4, 0)},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventAttach, effects.ThisBecameAttachedToACreature,
				"Enormous Energy Blade — tap that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !effects.OnBattlefield(g, item.Trigger.Event.Target) {
						return nil
					}
					return effects.TapTarget{Target: item.Trigger.Event.Target}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{2}")},
	})
}
