package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Captivating Unicorn — Creature — Unicorn {4}{W}, 4/4 (Foundations
// Jumpstart):
//
//	"Constellation — Whenever an enchantment you control enters, tap
//	 target creature an opponent controls."
//
// Constellation without the "or this creature" half: another enchantment
// entering under the Unicorn's controller (the Unicorn is not one, so
// its own entry never triggers). The target is chosen as the trigger goes
// on the stack; a target that left or stopped being legal does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "22005b01-84b1-4d4b-9193-3c561bd422ce",
		Name:         "Captivating Unicorn",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					c, ok := enteredUnderYourControl(ev, source, g, false)
					return ok && c.IsEnchantment()
				}, "Captivating Unicorn — tap target creature an opponent controls (constellation)", tapChosenPermanent),
				TargetCreature("target creature an opponent controls", OpponentControls()),
			),
		},
	})
}
