package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "7818e900-2f4e-4710-87a4-5a9ac170a677",
		Name:         "Enormous Energy Blade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(4, 0)},
		Triggered: []game.TriggeredAbility{
			On(game.EventAttach, ThisBecameAttachedToACreature,
				"Enormous Energy Blade — tap that creature",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil || !onBattlefield(g, item.Trigger.Event.Target) {
						return nil
					}
					return TapTarget{Target: item.Trigger.Event.Target}.Apply(NewContext(g, item))
				}),
		},
		Activated: []ActivatedAbility{EquipAbility("{2}")},
	})
}
