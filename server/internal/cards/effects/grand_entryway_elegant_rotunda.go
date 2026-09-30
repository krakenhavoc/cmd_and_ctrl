package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Room(RoomSpec{
		OracleID:     "f7574379-d10e-4bbf-b0b1-13feee753d8b",
		Name:         "Grand Entryway // Elegant Rotunda",
		Completeness: CompletenessFull,
		Left: Door{Triggered: []game.TriggeredAbility{
			WhenYouUnlockThisDoor(game.DoorLeft, "Grand Entryway — create a 1/1 white Glimmer enchantment creature token",
				Do(CreateToken{Template: TokenCard("1/1 white Glimmer enchantment"), N: 1})),
		}},
		Right: Door{Triggered: []game.TriggeredAbility{
			Targeting(
				WhenYouUnlockThisDoor(game.DoorRight, "Elegant Rotunda — put a +1/+1 counter on each of up to two target creatures",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures").WithCount(0, 2)),
		}},
	}))
}
