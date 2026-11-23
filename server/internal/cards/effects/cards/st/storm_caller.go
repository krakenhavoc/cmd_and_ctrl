package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Storm Caller — Creature — Ogre Shaman {2}{R}, 3/2 (Core Set 2021, #5511):
//
//	"When this creature enters, it deals 2 damage to each opponent."
//
// Fear of Burning Alive's enters damage at 2: one damage event per
// opponent, dealt by the creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5c946cc0-5d90-4970-acf8-b14eff177cbc",
		Name:         "Storm Caller",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Storm Caller — it deals 2 damage to each opponent", effects.Do(effects.DamageEachOpponent{N: 2})),
		},
	})
}
