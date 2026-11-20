package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Storm Caller — Creature — Ogre Shaman {2}{R}, 3/2 (Core Set 2021, #5511):
//
//	"When this creature enters, it deals 2 damage to each opponent."
//
// Fear of Burning Alive's enters damage at 2: one damage event per
// opponent, dealt by the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5c946cc0-5d90-4970-acf8-b14eff177cbc",
		Name:         "Storm Caller",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Storm Caller — it deals 2 damage to each opponent", Do(DamageEachOpponent{N: 2})),
		},
	})
}
