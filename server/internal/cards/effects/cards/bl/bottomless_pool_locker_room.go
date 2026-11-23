package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bottomless Pool // Locker Room — Enchantment — Room (CR 709.5):
//
//	Bottomless Pool {U}: "When you unlock this door, return up to one
//	target creature to its owner's hand."
//	Locker Room {4}{U}: "Whenever one or more creatures you control deal
//	combat damage to a player, draw a card."
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "b9ac4856-5cdb-479f-b0f5-7fa0b232a5ff",
		Name:         "Bottomless Pool // Locker Room",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{effects.Targeting(
			effects.WhenYouUnlockThisDoor(game.DoorLeft, "Bottomless Pool — return up to one target creature to its owner's hand",
				effects.BounceUpToOneTarget),
			effects.TargetCreature("up to one target creature").WithCount(0, 1),
		)}},
		Right: effects.Door{Triggered: []game.TriggeredAbility{
			effects.WheneverOneOrMoreCreaturesYouControlDealCombatDamageToAPlayer(nil,
				"Locker Room — draw a card", effects.Do(effects.DrawCards{N: 1})),
		}},
	}))
}
