package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// rooms_a_helpers.go — shared bodies for the first batch of Rooms
// (ADR 0103 PR 3).

// roomDamageToOpponentCreature is "When you unlock this door, this Room
// deals <amount> damage to target creature an opponent controls." The
// amount is read at resolution (Roaring Furnace counts the hand then).
func roomDamageToOpponentCreature(door game.DoorSide, label string, amount func(g *game.Game, item *game.StackItem) int) game.TriggeredAbility {
	return Targeting(
		WhenYouUnlockThisDoor(door, label, func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			ts := ctx.LegalTargets()
			if len(ts) == 0 {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: ts[0].ID, Amount: amount(g, item)}.Apply(ctx)
		}),
		TargetCreature("target creature an opponent controls", OpponentControls()),
	)
}
