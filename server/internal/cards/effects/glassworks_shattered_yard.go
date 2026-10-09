package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glassworks // Shattered Yard — Enchantment — Room (CR 709.5):
//
//	Glassworks {2}{R}: "When you unlock this door, this Room deals 4
//	damage to target creature an opponent controls."
//	Shattered Yard {4}{R}: "At the beginning of your end step, this Room
//	deals 1 damage to each opponent."
func init() {
	Register(Room(RoomSpec{
		OracleID:     "4122720b-cad9-4ebb-b458-be71411c21e3",
		Name:         "Glassworks // Shattered Yard",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{TriggerWithPurpose(roomDamageToOpponentCreature(game.DoorLeft,
			"Glassworks — 4 damage to target creature an opponent controls",
			func(*game.Game, *game.StackItem) int { return 4 }), ForTargets(DamageToTarget(0, 4)))}},
		Right: Door{Triggered: []game.TriggeredAbility{AtYourEndStep(
			"Shattered Yard — 1 damage to each opponent",
			func(g *game.Game, item *game.StackItem) error { return damageToEachOpponent(g, item, 1) })}},
	}))
}
