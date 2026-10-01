package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bottomless Pool // Locker Room — Enchantment — Room (CR 709.5):
//
//	Bottomless Pool {U}: "When you unlock this door, return up to one
//	target creature to its owner's hand."
//	Locker Room {4}{U}: "Whenever one or more creatures you control deal
//	combat damage to a player, draw a card."
func init() {
	Register(Room(RoomSpec{
		OracleID:     "b9ac4856-5cdb-479f-b0f5-7fa0b232a5ff",
		Name:         "Bottomless Pool // Locker Room",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{Targeting(
			WhenYouUnlockThisDoor(game.DoorLeft, "Bottomless Pool — return up to one target creature to its owner's hand",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return BounceToHand{Target: t.ID}.Apply(ctx)
					}
					return nil
				}),
			TargetCreature("up to one target creature").WithCount(0, 1),
		)}},
		Right: Door{Triggered: []game.TriggeredAbility{
			WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Locker Room — draw a card", Do(DrawCards{N: 1})),
		}},
	}))
}
