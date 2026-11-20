package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Where Ancients Tread — Enchantment {4}{R} (Shards of Alara, #5588):
//
//	"Whenever a creature you control with power 5 or greater enters, you
//	 may have this enchantment deal 5 damage to any target."
//
// Garruk's Packleader's trigger condition at power 5, read as the
// creature enters (CR 603.6a), with a target chosen as the trigger goes
// on the stack (CR 603.3d) and a "you may" asked as it resolves (CR
// 603.5). A target gone in response takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fca2fcab-4f17-448d-bf6d-f6c913159df8",
		Name:         "Where Ancients Tread",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				Optional(
					On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						c, ok := enteredUnderYourControl(ev, source, g, false)
						return ok && c.IsCreature() && c.CurrentPower() >= 5
					}, "Where Ancients Tread — 5 damage to any target", DamageToFirstTarget(5)),
					"Where Ancients Tread — deal 5 damage to that target?"),
				TargetAny()),
		},
	})
}
