package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Grand Entryway // Elegant Rotunda — Enchantment — Room (ADR 0103):
//
//	Grand Entryway {1}{W}: "When you unlock this door, create a 1/1
//	white Glimmer enchantment creature token."
//	Elegant Rotunda {2}{W}: "When you unlock this door, put a +1/+1
//	counter on each of up to two target creatures."
//
// The Rotunda's clause is Rishkar's ("each of up to two target
// creatures"), through the same shared resolution.
func init() {
	effects.Register(effects.Room(effects.RoomSpec{
		OracleID:     "f7574379-d10e-4bbf-b0b1-13feee753d8b",
		Name:         "Grand Entryway // Elegant Rotunda",
		Completeness: effects.CompletenessFull,
		Left: effects.Door{Triggered: []game.TriggeredAbility{
			effects.WhenYouUnlockThisDoor(game.DoorLeft, "Grand Entryway — create a 1/1 white Glimmer enchantment creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 white Glimmer enchantment"), N: 1})),
		}},
		Right: effects.Door{Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenYouUnlockThisDoor(game.DoorRight, "Elegant Rotunda — put a +1/+1 counter on each of up to two target creatures",
					effects.PutPlusOneCounterOnEachLegalTarget),
				effects.TargetCreature("up to two target creatures").WithCount(0, 2)),
		}},
	}))
}
